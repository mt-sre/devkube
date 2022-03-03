package dev

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-logr/logr"
)

type EnvironmentConfig struct {
	// Cluster initializers prepare a cluster for use.
	ClusterInitializers []ClusterInitializer
	// Container runtime to use
	ContainerRuntime ContainerRuntime
	Logger           logr.Logger
	NewCluster       NewClusterFunc
	ClusterOptions   []ClusterOption
}

// Apply default configuration.
func (c *EnvironmentConfig) Default() {
	if len(c.ContainerRuntime) == 0 {
		c.ContainerRuntime = ContainerRuntimeAuto
	}
	if c.Logger.GetSink() == nil {
		c.Logger = logr.Discard()
	}
	if c.NewCluster == nil {
		c.NewCluster = func(kubeconfigPath string, opts ...ClusterOption) (cluster, error) {
			return NewCluster(kubeconfigPath, opts...)
		}
	}

	// Prepend logger option to always default to the same logger for subcomponents.
	// Users can explicitly disable sub component logging by using:
	// WithLogger(logr.Discard()).
	c.ClusterOptions = append([]ClusterOption{
		WithLogger(c.Logger),
	}, c.ClusterOptions...)
}

type EnvironmentOption interface {
	ApplyToEnvironmentConfig(c *EnvironmentConfig)
}

type NewClusterFunc func(kubeconfigPath string, opts ...ClusterOption) (cluster, error)

type ClusterInitializer interface {
	Init(ctx context.Context, cluster cluster) error
}

// Environment represents a development environment.
type Environment struct {
	name string
	// Working directory of the environment.
	// Temporary files/kubeconfig etc. will be stored here.
	workDir string
	cluster cluster
	// container runtime in use by the environment
	// evaluated to "docker" or "podman" when config is set to "auto"
	containerRuntime ContainerRuntime
	config           EnvironmentConfig
}

// Creates a new development environment.
func NewEnvironment(name, workDir string, opts ...EnvironmentOption) *Environment {
	env := &Environment{
		name:    name,
		workDir: workDir,
	}
	for _, opt := range opts {
		opt.ApplyToEnvironmentConfig(&env.config)
	}
	env.config.Default()
	return env
}

// Initializes the environment and prepares it for use.
func (env *Environment) Init(ctx context.Context) error {
	// ensure workdir
	if err := os.MkdirAll(env.workDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating workdir: %w", err)
	}

	// determine container runtime
	if env.config.ContainerRuntime == ContainerRuntimeAuto {
		cr, err := DetermineContainerRuntime()
		if err != nil {
			return err
		}
		env.containerRuntime = cr
	} else {
		env.containerRuntime = env.config.ContainerRuntime
	}

	// Ensure the cluster is there
	cluster, err := env.ensureCluster(ctx)
	if err != nil {
		return fmt.Errorf("ensuring cluster: %w", err)
	}
	env.cluster = cluster

	return nil
}

func (env *Environment) WorkDir() string {
	return env.workDir
}

// Destroy/Teardown the development environment.
func (env *Environment) Destroy(ctx context.Context) error {
	if err := env.execKindCommand(
		ctx, os.Stdout, os.Stderr,
		"delete", "cluster",
		"--kubeconfig="+path.Join(env.workDir, "kubeconfig.yaml"),
		"--name="+env.name,
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
		"--name="+env.name,
	); err != nil {
		return fmt.Errorf("loading image archive: %w", err)
	}
	return nil
}

func (env *Environment) execKindCommand(
	ctx context.Context, stdout, stderr io.Writer, args ...string) error {
	env.config.Logger.Info("exec: kind " + strings.Join(args, " "))
	kindCmd := exec.CommandContext( //nolint:gosec
		ctx, "kind", args...,
	)
	kindCmd.Env = os.Environ()
	if env.containerRuntime == "podman" {
		kindCmd.Env = append(kindCmd.Env, "KIND_EXPERIMENTAL_PROVIDER=podman")
	}
	kindCmd.Stdout = stdout
	kindCmd.Stderr = stderr
	return kindCmd.Run()
}

func (env *Environment) ensureCluster(ctx context.Context) (cluster, error) {
	kubeconfigPath := path.Join(env.workDir, "kubeconfig.yaml")

	clusterExists, err := env.kindClusterExist(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !clusterExists {
		kindConfigPath, err := env.writeKindConfig()
		if err != nil {
			return nil, fmt.Errorf("writing kind config: %w", err)
		}

		// Create cluster
		if err := env.execKindCommand(
			ctx, os.Stdout, os.Stderr,
			"create", "cluster",
			"--kubeconfig="+kubeconfigPath,
			"--name="+env.name,
			"--config="+kindConfigPath,
		); err != nil {
			return nil, fmt.Errorf("creating kind cluster: %w", err)
		}
	}

	// Create _all_ the clients
	cluster, err := env.config.NewCluster(
		env.workDir, append(env.config.ClusterOptions, WithKubeconfigPath(kubeconfigPath))...)
	if err != nil {
		return nil, fmt.Errorf("creating k8s clients: %w", err)
	}
	env.cluster = cluster

	// Run ClusterInitializers
	if !clusterExists {
		for _, initializer := range env.config.ClusterInitializers {
			if err := initializer.Init(ctx, cluster); err != nil {
				return nil, fmt.Errorf("running cluster initializer: %w", err)
			}
		}
	}
	return cluster, nil
}

func (env *Environment) kindClusterExist(ctx context.Context) (bool, error) {
	var checkOutput bytes.Buffer
	if err := env.execKindCommand(ctx, &checkOutput, os.Stderr, "get", "clusters"); err != nil {
		return false, fmt.Errorf("getting existing kind clusters: %w\n%s", err, checkOutput.String())
	}
	return strings.Contains(checkOutput.String(), env.name+"\n"), nil
}

const kindConfigFile = "kind.yaml"

func (env *Environment) writeKindConfig() (kindConfigPath string, err error) {
	kindConfigPath = path.Join(env.workDir, kindConfigFile)
	kindConfig := `kind: Cluster
		apiVersion: kind.x-k8s.io/v1alpha4
`

	if env.containerRuntime == ContainerRuntimePodman {
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
	}
	if err := ioutil.WriteFile(
		kindConfigPath, []byte(kindConfig), os.ModePerm); err != nil {
		return "", fmt.Errorf("creating kind cluster config: %w", err)
	}
	return kindConfigPath, nil
}

// Try to figure out the installed container runtime from the environment.
func DetermineContainerRuntime() (ContainerRuntime, error) {
	// check for docker
	if _, err := exec.LookPath("docker"); err == nil {
		return ContainerRuntimeDocker, nil
	} else if err != nil &&
		!errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("looking up docker binary: %w", err)
	}

	// check for podman
	if _, err := exec.LookPath("podman"); err == nil {
		return ContainerRuntimePodman, nil
	} else if err != nil &&
		!errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("looking up podman binary: %w", err)
	}

	return "", fmt.Errorf("no container runtime found, tried: docker, podman")
}
