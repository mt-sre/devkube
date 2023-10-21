package devhelm

import (
	"context"
	"fmt"
)

type InstallArg func(*[]string)

func InstallWithNamespace(namespace string) InstallArg {
	return func(s *[]string) {
		*s = append(*s, "--namespace", namespace)
	}
}

func InstallWithSet(set string) InstallArg {
	return func(s *[]string) {
		*s = append(*s, "--set", set)
	}
}

func (h RealHelm) Install(ctx context.Context, name, chart string, opts ...InstallArg) error {
	args := []string{"install", name, chart}
	for _, o := range opts {
		o(&args)
	}

	err := h.ExecHelmCmd(ctx, args)
	if err != nil {
		return fmt.Errorf("helm repo update: %w", err)
	}
	return nil
}
