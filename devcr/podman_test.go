package devcr_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mt-sre/devkube/devcr"
	"github.com/mt-sre/devkube/devos"
)

func TestPodmanErrReporting(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Twice()
	cmd.On("SetStdin", os.Stdin).Twice()
	cmd.On("SetStdout", os.Stdout).Twice()

	exec.On("CommandContext", mock.Anything, mock.Anything, mock.Anything).Return(cmd).Once()
	cmd.On("Run").Once().Return(nil)
	err := p.BuildImage(ctx, tag, push, srcPath, file)
	require.NoError(t, err)

	errIn := errors.New("cheeeese")
	exec.On("CommandContext", mock.Anything, mock.Anything, mock.Anything).Return(cmd).Once()
	cmd.On("Run").Once().Return(errIn)
	err = p.BuildImage(ctx, tag, push, srcPath, file)
	require.ErrorIs(t, err, errIn)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestPodmanBuildImage(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Times(4)
	cmd.On("SetStdin", os.Stdin).Times(4)
	cmd.On("SetStdout", os.Stdout).Times(4)
	cmd.On("Run").Return(nil).Times(4)

	exec.On("CommandContext", ctx, "podman", []string{"build", "srcPath"}).Return(cmd).Once()
	require.NoError(t, p.BuildImage(ctx, "", false, srcPath, ""))

	exec.On("CommandContext", ctx, "podman", []string{"build", "--tag", "tag", "srcPath"}).Return(cmd).Once()
	require.NoError(t, p.BuildImage(ctx, "tag", false, srcPath, ""))

	exec.On("CommandContext", ctx, "podman", []string{"build", "--tag", "tag", "--push", "srcPath"}).Return(cmd).Once()
	require.NoError(t, p.BuildImage(ctx, "tag", true, srcPath, ""))

	exec.On("CommandContext", ctx, "podman", []string{"build", "--file", "file", "srcPath"}).Return(cmd).Once()
	require.NoError(t, p.BuildImage(ctx, "", false, srcPath, "file"))

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestPodmanLoadImage(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("Run").Return(nil).Once()

	exec.On("CommandContext", ctx, "podman", []string{"load", "--input", srcPath}).Return(cmd).Once()
	require.NoError(t, p.LoadImage(ctx, srcPath))
}

func TestPodmanLoginAtImageRegistry(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("Run").Return(nil).Once()

	exec.On(
		"CommandContext", ctx, "podman",
		[]string{"login", "--username", user, "--password", password, registry},
	).Return(cmd).Once()
	require.NoError(t, p.LoginAtImageRegistry(ctx, registry, user, password))
}

func TestPodmanPushImage(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("Run").Return(nil).Once()

	exec.On("CommandContext", ctx, "podman", []string{"push", tag}).Return(cmd).Once()
	require.NoError(t, p.PushImage(ctx, tag))
}

func TestPodmanSaveImage(t *testing.T) {
	t.Parallel()

	exec := &devos.MockExec{}
	p := devcr.Podman{exec}
	ctx := context.Background()

	tags := []string{"a", "b"}

	cmd := &devos.MockCmd{}
	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("Run").Return(nil).Once()

	exec.On("CommandContext", ctx, "podman", []string{"save", "--output", dstPath, tags[0], tags[1]}).Return(cmd).Once()
	require.NoError(t, p.SaveImage(ctx, dstPath, tags...))
}
