package devcr

import (
	"context"

	"github.com/mt-sre/devkube/devos"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

type Podman struct{ Exec devos.Exec }

// BuildImage builds an image via the runtime.
//
// If parameter tag is set the built image gets this tag assigned.
// If push is true the resulting image gets pushed after it is tagged.
// SrcPath defines the directory the build runs on.
// If file is set the Containerfile if expected at this path.
func (p Podman) BuildImage(ctx context.Context, tag string, push bool, srcPath, file string) error {
	return execute(ctx, p.Exec, "podman", dockerPodmanBuildArgs(tag, push, srcPath, file))
}

// LoadImage loads the image contained in the file located at parameter srcPath into
// the container runtime.
func (p Podman) LoadImage(ctx context.Context, srcPath string) error {
	return execute(ctx, p.Exec, "podman", dockerPodmanLoadArgs(srcPath))
}

// SaveImage lets the runtime save existing images with the given tags into
// an archive located at dstPath.
func (p Podman) SaveImage(ctx context.Context, dstPath string, tag ...string) error {
	return execute(ctx, p.Exec, "podman", dockerPodmanSaveArgs(dstPath, tag))
}

// LoginAtImageRegistry lets the runtime perform a login against the specified
// registry with the given user and password.
func (p Podman) LoginAtImageRegistry(ctx context.Context, registry, user, password string) error {
	return execute(ctx, p.Exec, "podman", dockerPodmanLoginArgs(registry, user, password))
}

// KindProvider returns the appropriate kind provider for the specific runtime.
func (p Podman) KindProvider() *kindcluster.Provider {
	return kindcluster.NewProvider(
		kindcluster.ProviderWithLogger(kindcmd.NewLogger()),
		kindcluster.ProviderWithPodman(),
	)
}

// PushImage lets the runtime push an existing image tag.
func (p Podman) PushImage(ctx context.Context, tag string) error {
	return execute(ctx, p.Exec, "podman", dockerPodmanPushArgs(tag))
}

func (p Podman) test(ctx context.Context) error {
	return p.Exec.CommandContext(ctx, "podman", "info").Run()
}
