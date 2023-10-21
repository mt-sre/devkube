package devcr

import (
	"context"

	"github.com/mt-sre/devkube/devos"

	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

type Docker struct{ Exec devos.Exec }

// BuildImage builds an image via the runtime.
//
// If parameter tag is set the built image gets this tag assigned.
// If push is true the resulting image gets pushed after it is tagged.
// SrcPath defines the directory the build runs on.
// If file is set the Containerfile if expected at this path.
func (d Docker) BuildImage(ctx context.Context, tag string, push bool, srcPath, file string) error {
	return execute(ctx, d.Exec, "docker", dockerPodmanBuildArgs(tag, push, srcPath, file))
}

// LoadImage loads the image contained in the file located at parameter srcPath into
// the container runtime.
func (d Docker) LoadImage(ctx context.Context, srcPath string) error {
	return execute(ctx, d.Exec, "docker", dockerPodmanLoadArgs(srcPath))
}

// SaveImage lets the runtime save existing images with the given tags into
// an archive located at dstPath.
func (d Docker) SaveImage(ctx context.Context, dstPath string, tags ...string) error {
	return execute(ctx, d.Exec, "docker", dockerPodmanSaveArgs(dstPath, tags))
}

// LoginAtImageRegistry lets the runtime perform a login against the specified
// registry with the given user and password.
func (d Docker) LoginAtImageRegistry(ctx context.Context, registry, user, password string) error {
	return execute(ctx, d.Exec, "docker", dockerPodmanLoginArgs(registry, user, password))
}

// KindProvider returns the appropriate kind provider for the specific runtime.
func (d Docker) KindProvider() *kindcluster.Provider {
	return kindcluster.NewProvider(kindcluster.ProviderWithLogger(kindcmd.NewLogger()), kindcluster.ProviderWithDocker())
}

// PushImage lets the runtime push an existing image tag.
func (d Docker) PushImage(ctx context.Context, tag string) error {
	return execute(ctx, d.Exec, "docker", dockerPodmanPushArgs(tag))
}

func (d Docker) test(ctx context.Context) error {
	return d.Exec.CommandContext(ctx, "docker", "info").Run()
}
