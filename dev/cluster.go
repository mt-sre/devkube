package dev

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var defaultSchemeBuilder runtime.SchemeBuilder = runtime.SchemeBuilder{
	clientgoscheme.AddToScheme,
	apiextensionsv1.AddToScheme,
}

type ClusterConfig struct {
	Logger        logr.Logger
	SchemeBuilder runtime.SchemeBuilder
	NewWaiter     NewWaiterFunc
	WaitOptions   []WaitOption
	NewHelm       NewHelmFunc
	HelmOptions   []HelmOption
	NewRestConfig NewRestConfigFunc
	NewCtrlClient NewCtrlClientFunc

	WorkDir string
	// Path to the kubeconfig of the cluster
	Kubeconfig string
}

type NewWaiterFunc func(
	client client.Client, scheme *runtime.Scheme,
	defaultOpts ...WaitOption,
) *Waiter

type NewHelmFunc func(
	workDir, kubeconfig string,
	opts ...HelmOption,
) helm

type NewRestConfigFunc func(kubeconfig string) (*rest.Config, error)

func DefaultNewRestConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

type NewCtrlClientFunc func(c *rest.Config, opts client.Options) (client.Client, error)

var DefaultNewCtrlClientFunc = client.New

func (c *ClusterConfig) Default() {
	if c.Logger.GetSink() == nil {
		c.Logger = logr.Discard()
	}
	if c.NewWaiter == nil {
		c.NewWaiter = NewWaiter
	}
	if c.NewHelm == nil {
		c.NewHelm = func(
			workDir, kubeconfig string,
			opts ...HelmOption,
		) helm {
			return NewHelm(workDir, kubeconfig, opts...)
		}
	}
	if c.Kubeconfig == "" {
		c.Kubeconfig = path.Join(c.WorkDir, "kubeconfig.yaml")
	}
	if c.NewRestConfig == nil {
		c.NewRestConfig = DefaultNewRestConfig
	}
	if c.NewCtrlClient == nil {
		c.NewCtrlClient = DefaultNewCtrlClientFunc
	}

	// Prepend logger option to always default to the same logger for subcomponents.
	// Users can explicitly disable sub component logging by using:
	// WithLogger(logr.Discard()).
	c.WaitOptions = append([]WaitOption{
		WithLogger(c.Logger),
	}, c.WaitOptions...)
}

type ClusterOption interface {
	ApplyToClusterConfig(c *ClusterConfig)
}

type cluster interface {
	CreateAndWaitFromHttp(
		ctx context.Context, urls []string,
		opts ...WaitOption,
	) error
	CreateAndWaitFromFiles(
		ctx context.Context, files []string,
		opts ...WaitOption,
	) error
	CreateAndWaitFromFolders(
		ctx context.Context, folders []string,
		opts ...WaitOption,
	) error
	CreateAndWaitForReadiness(
		ctx context.Context, object client.Object,
		opts ...WaitOption,
	) error
	Kubeconfig() string
	Helm() helm
	CtrlClient() client.Client
}

var _ cluster = (*Cluster)(nil)

// Container object to hold kubernetes client interfaces and configuration.
type Cluster struct {
	scheme     *runtime.Scheme
	restConfig *rest.Config
	ctrlClient client.Client
	waiter     *Waiter
	helm       helm

	config ClusterConfig
}

// Creates a new Cluster object to interact with a Kubernetes cluster.
func NewCluster(workDir string, opts ...ClusterOption) (*Cluster, error) {
	c := &Cluster{
		scheme: runtime.NewScheme(),
		config: ClusterConfig{
			WorkDir: workDir,
		},
	}

	// Add default schemes
	if err := defaultSchemeBuilder.AddToScheme(c.scheme); err != nil {
		return nil, fmt.Errorf("adding defaults to scheme: %w", err)
	}

	// Apply Options
	for _, opt := range opts {
		opt.ApplyToClusterConfig(&c.config)
	}
	c.config.Default()

	// Apply schemes from Options
	if c.config.SchemeBuilder != nil {
		if err := c.config.SchemeBuilder.AddToScheme(c.scheme); err != nil {
			return nil, fmt.Errorf("adding to scheme: %w", err)
		}
	}

	var err error
	// Create RestConfig
	c.restConfig, err = c.config.NewRestConfig(c.config.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("getting rest.Config from kubeconfig: %w", err)
	}

	// Create Controller Runtime Client
	c.ctrlClient, err = c.config.NewCtrlClient(c.restConfig, client.Options{
		Scheme: c.scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("creating new ctrl client: %w", err)
	}

	c.waiter = c.config.NewWaiter(
		c.ctrlClient, c.scheme,
		c.config.WaitOptions...)
	c.helm = c.config.NewHelm(
		workDir, c.config.Kubeconfig,
		c.config.HelmOptions...)

	return c, nil
}

func (c *Cluster) Helm() helm {
	return c.helm
}

func (c *Cluster) CtrlClient() client.Client {
	return c.ctrlClient
}

// Returns the path to the kubeconfig of the cluster.
func (c *Cluster) Kubeconfig() string {
	return c.config.Kubeconfig
}

// Load kube objects from a list of http urls,
// create these objects and wait for them to be ready.
func (c *Cluster) CreateAndWaitFromHttp(
	ctx context.Context, urls []string,
	opts ...WaitOption,
) error {
	var client http.Client
	var objects []unstructured.Unstructured
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("getting %q: %w", url, err)
		}
		defer resp.Body.Close()

		var content bytes.Buffer
		if _, err := io.Copy(&content, resp.Body); err != nil {
			return fmt.Errorf("reading response %q: %w", url, err)
		}

		objs, err := LoadKubernetesObjectsFromBytes(content.Bytes())
		if err != nil {
			return fmt.Errorf("loading objects from %q: %w", url, err)
		}

		objects = append(objects, objs...)
	}

	for i := range objects {
		if err := c.CreateAndWaitForReadiness(ctx, &objects[i], opts...); err != nil {
			return fmt.Errorf("creating object: %w", err)
		}
	}
	return nil
}

// Load kube objects from a list of files,
// create these objects and wait for them to be ready.
func (c *Cluster) CreateAndWaitFromFiles(
	ctx context.Context, files []string,
	opts ...WaitOption,
) error {
	var objects []unstructured.Unstructured
	for _, file := range files {
		objs, err := LoadKubernetesObjectsFromFile(file)
		if err != nil {
			return fmt.Errorf("loading objects from file %q: %w", file, err)
		}

		objects = append(objects, objs...)
	}

	for i := range objects {
		if err := c.CreateAndWaitForReadiness(ctx, &objects[i], opts...); err != nil {
			return fmt.Errorf("creating object: %w", err)
		}
	}
	return nil
}

// Load kube objects from a list of folders,
// create these objects and wait for them to be ready.
func (c *Cluster) CreateAndWaitFromFolders(
	ctx context.Context, folders []string,
	opts ...WaitOption,
) error {
	var objects []unstructured.Unstructured
	for _, folder := range folders {
		objs, err := LoadKubernetesObjectsFromFolder(folder)
		if err != nil {
			return fmt.Errorf("loading objects from folder %q: %w", folder, err)
		}

		objects = append(objects, objs...)
	}

	for i := range objects {
		if err := c.CreateAndWaitForReadiness(ctx, &objects[i], opts...); err != nil {
			return fmt.Errorf("creating object: %w", err)
		}
	}
	return nil
}

// Creates the given objects and waits for them to be considered ready.
func (c *Cluster) CreateAndWaitForReadiness(
	ctx context.Context, object client.Object,
	opts ...WaitOption,
) error {
	if err := c.ctrlClient.Create(ctx, object); err != nil &&
		!errors.IsAlreadyExists(err) {
		return fmt.Errorf("creating object: %w", err)
	}

	if err := c.waiter.WaitForReadiness(ctx, object); err != nil {
		var unknownTypeErr *UnknownTypeError
		if goerrors.As(err, &unknownTypeErr) {
			// A lot of types don't require waiting for readiness,
			// so we should not error in cases when object types
			// are not registered for the generic wait method.
			return nil
		}

		return fmt.Errorf("waiting for object: %w", err)
	}
	return nil
}
