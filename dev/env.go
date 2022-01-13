package dev

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

type EnvironmentConfig struct {
	// Cluster initializers prepare a cluster for use.
	ClusterInitializers []ClusterInitializer
	// Container runtime to use
	ContainerRuntime string
	Logger           logr.Logger
	NewCluster       NewClusterFunc
	ClusterOptions   []ClusterOption
}

// Apply default configuration.
func (c *EnvironmentConfig) Default() {
	if len(c.ContainerRuntime) == 0 {
		c.ContainerRuntime = "podman"
	}
	if c.Logger.GetSink() == nil {
		c.Logger = logr.Discard()
	}
	if c.NewCluster == nil {
		c.NewCluster = NewCluster
	}
	if c.ClusterOptions == nil {
		c.ClusterOptions = append(c.ClusterOptions, ClusterWithLogger(c.Logger))
	}
}

type EnvironmentOption func(c *EnvironmentConfig)

func EnvironmentWithClusterInitializers(init ...ClusterInitializer) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.ClusterInitializers = append(c.ClusterInitializers, init...)
	}
}

func EnvironmentWithContainerRuntime(containerRuntime string) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.ContainerRuntime = containerRuntime
	}
}

func EnvironmentWithLogger(logger logr.Logger) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.Logger = logger
	}
}

func EnvironmentWithNewClusterFunc(newClusterFn NewClusterFunc) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.NewCluster = newClusterFn
	}
}

func EnvironmentWithClusterOptions(opts ...ClusterOption) EnvironmentOption {
	return func(c *EnvironmentConfig) {
		c.ClusterOptions = opts
	}
}

type NewClusterFunc func(kubeconfigPath string, opts ...ClusterOption) (*Cluster, error)

type ClusterInitializer interface {
	Init(ctx context.Context, cluster *Cluster) error
}

// Load objects from given folder paths and applies them into the cluster.
type ClusterLoadObjectsFromFolder []string

func (l ClusterLoadObjectsFromFolder) Init(
	ctx context.Context, cluster *Cluster) error {
	return cluster.CreateAndWaitFromFolders(ctx, l)
}

// Load objects from given file paths and applies them into the cluster.
type ClusterLoadObjectsFromFiles []string

func (l ClusterLoadObjectsFromFiles) Init(
	ctx context.Context, cluster *Cluster) error {
	return cluster.CreateAndWaitFromFiles(ctx, l)
}

type ClusterLoadObjectsFromHttp []string

func (l ClusterLoadObjectsFromHttp) Init(
	ctx context.Context, cluster *Cluster) error {
	return cluster.CreateAndWaitFromHttp(ctx, l)
}

type ClusterHelmInstall struct {
	RepoName, RepoURL, PackageName, Namespace string
	SetVars                                   []string
}

func (l ClusterHelmInstall) Init(
	ctx context.Context, cluster *Cluster) error {
	return cluster.HelmInstall(
		ctx, cluster,
		l.RepoName, l.RepoURL, l.PackageName, l.Namespace, l.SetVars)
}

// Environment represents a development environment.
type Environment struct {
	Name string
	// Working directory of the environment.
	// Temporary files/kubeconfig etc. will be stored here.
	WorkDir string
	Cluster *Cluster
	config  EnvironmentConfig
}

// Creates a new development environment.
func NewEnvironment(name, workDir string, opts ...EnvironmentOption) *Environment {
	env := &Environment{
		Name:    name,
		WorkDir: workDir,
	}
	for _, opt := range opts {
		opt(&env.config)
	}
	env.config.Default()
	return env
}

// Initializes the environment and prepares it for use.
func (env *Environment) Init(ctx context.Context) error {
	kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
`

	// Workaround for https://github.com/kubernetes-sigs/kind/issues/2411
	// For BTRFS on LUKS.
	if _, err := os.Lstat("/dev/dm-0"); err == nil {
		kindConfig += `nodes:
- role: control-plane
  extraMounts:
    - hostPath: /dev/dm-0
      containerPath: /dev/dm-0
      propagation: HostToContainer
`
	}

	if err := os.MkdirAll(env.WorkDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating workdir: %w", err)
	}

	kindConfigPath := env.WorkDir + "/kind.yaml"
	if err := ioutil.WriteFile(
		kindConfigPath, []byte(kindConfig), os.ModePerm); err != nil {
		return fmt.Errorf("creating kind cluster config: %w", err)
	}

	// Needs cluster creation?
	var checkOutput bytes.Buffer
	if err := env.execKindCommand(ctx, &checkOutput, nil, "get", "clusters"); err != nil {
		return fmt.Errorf("getting existing kind clusters: %w", err)
	}

	// Only create cluster if it is not already there.
	createCluster := !strings.Contains(checkOutput.String(), env.Name+"\n")

	if createCluster {
		// Create cluster
		if err := env.execKindCommand(
			ctx, os.Stdout, os.Stderr,
			"create", "cluster",
			"--kubeconfig="+env.Cluster.Kubeconfig, "--name="+env.Name,
			"--config="+kindConfigPath,
		); err != nil {
			return fmt.Errorf("creating kind cluster: %w", err)
		}
	}

	// Create _all_ the clients
	cluster, err := NewCluster(
		env.WorkDir+"/kubeconfig.yaml", env.config.ClusterOptions...)
	if err != nil {
		return fmt.Errorf("creating k8s clients: %w", err)
	}
	env.Cluster = cluster

	// Run ClusterInitializers
	if createCluster {
		for _, initializer := range env.config.ClusterInitializers {
			if err := initializer.Init(ctx, cluster); err != nil {
				return fmt.Errorf("running cluster initializer: %w", err)
			}
		}
	}

	return nil
}

// Destroy/Teardown the development environment.
func (env *Environment) Destroy(ctx context.Context) error {
	if err := env.execKindCommand(
		ctx, os.Stdout, os.Stderr,
		"delete", "cluster",
		"--kubeconfig="+env.Cluster.Kubeconfig, "--name="+env.Name,
	); err != nil {
		return fmt.Errorf("deleting kind cluster: %w", err)
	}
	return nil
}

// Load an image from a tar archive into the environment.
func (env *Environment) LoadImageFromTar(
	ctx context.Context, filePath string) error {
	if err := env.execKindCommand(
		ctx, os.Stdout, os.Stderr,
		"load", "image-archive", filePath,
		"--name="+env.Name,
	); err != nil {
		return fmt.Errorf("loading image archive: %w", err)
	}
	return nil
}

func (env *Environment) execKindCommand(
	ctx context.Context, stdout, stderr io.Writer, args ...string) error {
	kindCmd := exec.CommandContext( //nolint:gosec
		ctx, "kind", args...,
	)
	kindCmd.Env = os.Environ()
	if env.config.ContainerRuntime == "podman" {
		kindCmd.Env = append(kindCmd.Env, "KIND_EXPERIMENTAL_PROVIDER=podman")
	}
	kindCmd.Stdout = stdout
	kindCmd.Stderr = stderr
	return kindCmd.Run()
}
