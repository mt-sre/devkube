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
	repoName = "repoName"
	repoURL  = "repoURL"
)

func TestRepoAddSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"repo", "add", repoName, repoURL}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.RepoAdd(ctx, repoName, repoURL)
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestRepoAddError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runEerr := errors.New("runErr")

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(runEerr)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"repo", "add", repoName, repoURL}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.RepoAdd(ctx, repoName, repoURL)
	require.ErrorIs(t, err, runEerr)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestRepoUpdateSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"repo", "update"}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.RepoUpdate(ctx)
	require.ErrorIs(t, err, nil)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestRepoUpdateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runEerr := errors.New("runErr")

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(runEerr)

	exec := &devos.MockExec{}
	exec.On("CommandContext", ctx, "helm", []string{"repo", "update"}).Once().Return(cmd)
	helm := devhelm.RealHelm{Exec: exec}

	err := helm.RepoUpdate(ctx)
	require.ErrorIs(t, err, runEerr)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}
