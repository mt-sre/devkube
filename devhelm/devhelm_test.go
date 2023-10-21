package devhelm_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mt-sre/devkube/devhelm"
	"github.com/mt-sre/devkube/devos"
)

func TestExecHelmCmdAllEnvSet(t *testing.T) {
	t.Parallel()

	helm := devhelm.RealHelm{
		KubeconfigPath: "kubeconfigPath",
		CacheDir:       "cacheDir",
		ConfigDir:      "configDir",
		DataDir:        "dataDir",
	}

	cmd := &devos.MockCmd{}
	cmd.On("SetEnv", []string{
		fmt.Sprintf("KUBECONFIG=%s", helm.KubeconfigPath),
		fmt.Sprintf("HELM_CACHE_HOME=%s", helm.CacheDir),
		fmt.Sprintf("HELM_CONFIG_HOME=%s", helm.ConfigDir),
		fmt.Sprintf("HELM_DATA_HOME=%s", helm.DataDir),
	}).Once()

	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	ctx := context.Background()
	exec := &devos.MockExec{}
	helm.Exec = exec
	args := []string{"a", "b", "c"}
	exec.On("CommandContext", ctx, "helm", args).Once().Return(cmd)

	err := helm.ExecHelmCmd(ctx, args)
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}

func TestExecHelmCmdNoEnvSet(t *testing.T) {
	t.Parallel()

	cmd := &devos.MockCmd{}

	cmd.On("SetStderr", os.Stderr).Once()
	cmd.On("SetStdout", os.Stdout).Once()
	cmd.On("SetStdin", os.Stdin).Once()
	cmd.On("Run").Once().Return(nil)

	ctx := context.Background()
	exec := &devos.MockExec{}
	helm := devhelm.RealHelm{Exec: exec}
	args := []string{"a", "b", "c"}
	exec.On("CommandContext", ctx, "helm", args).Once().Return(cmd)

	err := helm.ExecHelmCmd(ctx, args)
	require.NoError(t, err)

	exec.AssertExpectations(t)
	cmd.AssertExpectations(t)
}
