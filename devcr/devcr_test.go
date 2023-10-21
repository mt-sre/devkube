package devcr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/devkube/devcr"
	"github.com/mt-sre/devkube/devos"
)

func TestKindProvider(t *testing.T) {
	t.Parallel()

	p := devcr.Podman{}
	d := devcr.Docker{}

	providerP := p.KindProvider()
	providerD := d.KindProvider()
	require.NotNil(t, providerD)
	require.NotNil(t, providerP)
	require.NotEqual(t, providerD, providerP)
}

func TestDetect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	exec := &devos.MockExec{}
	podmanCmd := &devos.MockCmd{}
	dockerCmd := &devos.MockCmd{}

	exec.On("CommandContext", ctx, "podman", []string{"info"}).Once().Return(podmanCmd).Once()
	podmanCmd.On("Run").Return(nil).Once()
	cr, err := devcr.Detect(ctx, exec)
	require.NoError(t, err)
	require.IsType(t, devcr.Podman{}, cr)

	exec.On("CommandContext", ctx, "docker", []string{"info"}).Once().Return(dockerCmd).Once()
	exec.On("CommandContext", ctx, "podman", []string{"info"}).Once().Return(podmanCmd).Once()
	podmanErr := errors.New("nopodman")
	podmanCmd.On("Run").Return(podmanErr).Once()
	dockerCmd.On("Run").Return(nil).Once()
	cr, err = devcr.Detect(ctx, exec)
	require.NoError(t, err)
	require.IsType(t, devcr.Docker{}, cr)

	exec.On("CommandContext", ctx, "docker", []string{"info"}).Once().Return(dockerCmd).Once()
	exec.On("CommandContext", ctx, "podman", []string{"info"}).Once().Return(podmanCmd).Once()
	dockerErr := errors.New("nopodman")
	podmanCmd.On("Run").Return(podmanErr).Once()
	dockerCmd.On("Run").Return(dockerErr).Once()
	cr, err = devcr.Detect(ctx, exec)
	require.ErrorIs(t, err, dockerErr)
	require.ErrorIs(t, err, podmanErr)
	require.Nil(t, cr)
}
