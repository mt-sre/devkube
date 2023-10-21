package devhelm

import (
	"context"
	"fmt"
)

func (h RealHelm) RepoAdd(ctx context.Context, repoName, repoURL string) error {
	err := h.ExecHelmCmd(ctx, []string{"repo", "add", repoName, repoURL})
	if err != nil {
		return fmt.Errorf("helm repo add: %w", err)
	}
	return nil
}

func (h RealHelm) RepoUpdate(ctx context.Context) error {
	err := h.ExecHelmCmd(ctx, []string{"repo", "update"})
	if err != nil {
		return fmt.Errorf("helm repo update: %w", err)
	}
	return nil
}
