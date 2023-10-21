package devhelm_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/devkube/devhelm"
	"github.com/mt-sre/devkube/devos"
)

const (
	chart     = "chart"
	name      = "name"
	namespace = "namespace"
	set       = "set"
)

func TestInstallError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runEerr := errors.New("runErr")

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(runEerr)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"install", name, chart}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.Install(ctx, name, chart)
	require.ErrorIs(t, err, runEerr)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestInstallSucess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"install", name, chart}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.Install(ctx, name, chart)
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestInstallNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"install", name, chart, "--namespace", namespace}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.Install(ctx, name, chart, devhelm.InstallWithNamespace(namespace))
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestInstallSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"install", name, chart, "--set", set}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.Install(ctx, name, chart, devhelm.InstallWithSet(set))
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestInstallAllArgs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx,
		"helm", []string{"install", name, chart, "--namespace", namespace, "--set", set},
	).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.Install(ctx, name, chart, devhelm.InstallWithNamespace(namespace), devhelm.InstallWithSet(set))
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}
