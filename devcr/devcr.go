// Package devcr allows interacting with container runtimes.
package devcr

import (
	"context"
	"errors"
	"fmt"
	"os"

	"k8s.io/utils/exec"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"

	"github.com/mt-sre/devkube/devos"
)

// ContainerRuntime allows operation on the container runtime on the hosting system.
type ContainerRuntime interface {
	// BuildImage builds an image via the runtime.
	//
	// If parameter tag is set the built image gets this tag assigned.
	// If push is true the resulting image gets pushed after it is tagged.
	// SrcPath defines the directory the build runs on.
	// If file is set the Containerfile if expected at this path.
	BuildImage(ctx context.Context, tag string, push bool, srcPath, file string) error
	// KindProvider returns the appropriate kind provider for the specific runtime.
	KindProvider() *kindcluster.Provider
	// LoadImage loads the image contained in the file located at parameter srcPath into
	// the container runtime.
	LoadImage(ctx context.Context, srcPath string) error
	// LoginAtImageRegistry lets the runtime perform a login against the specified
	// registry with the given user and password.
	LoginAtImageRegistry(ctx context.Context, registry, user, password string) error
	// PushImage lets the runtime push an existing image tag.
	PushImage(ctx context.Context, tag string) error
	// SaveImage lets the runtime save existing images with the given tags into
	// an archive located at dstPath.
	SaveImage(ctx context.Context, dstPath string, tags ...string) error
}

func Detect(ctx context.Context, exec exec.Interface) (ContainerRuntime, error) {
	p := Podman{exec}
	pErr := p.test(ctx)
	if pErr == nil {
		return p, nil
	}
	d := Docker{exec}
	dErr := d.test(ctx)
	if dErr == nil {
		return d, nil
	}

	return nil, fmt.Errorf("no container runtime available: %w", errors.Join(pErr, dErr))
}

func execute(ctx context.Context, e exec.Interface, cmd string, args []string) error {
	e = devos.RealExecIfUnset(e)
	c := e.CommandContext(ctx, cmd, args...)
	c.SetStderr(os.Stderr)
	c.SetStdin(os.Stdin)
	c.SetStdout(os.Stdout)

	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", cmd, args, err)
	}

	return nil
}
