// Package devhelm manages package installations via helm.
package devhelm

import (
	"context"
	"os"

	"k8s.io/utils/exec"

	"github.com/mt-sre/devkube/devos"
)

type RealHelm struct {
	Exec           exec.Interface
	KubeconfigPath string
	CacheDir       string
	ConfigDir      string
	DataDir        string
}

func (h RealHelm) ExecHelmCmd(ctx context.Context, args []string) error {
	env := []string{}
	if h.KubeconfigPath != "" {
		env = append(env, "KUBECONFIG="+h.KubeconfigPath)
	}
	if h.CacheDir != "" {
		env = append(env, "HELM_CACHE_HOME="+h.CacheDir)
	}
	if h.ConfigDir != "" {
		env = append(env, "HELM_CONFIG_HOME="+h.ConfigDir)
	}
	if h.DataDir != "" {
		env = append(env, "HELM_DATA_HOME="+h.DataDir)
	}

	cmd := devos.RealExecIfUnset(h.Exec).CommandContext(ctx, "helm", args...)

	if len(env) > 0 {
		cmd.SetEnv(env)
	}
	cmd.SetStderr(os.Stderr)
	cmd.SetStdout(os.Stdout)
	cmd.SetStdin(os.Stdin)

	return cmd.Run()
}
