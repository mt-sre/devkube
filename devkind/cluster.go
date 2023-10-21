package devkind

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"

	"github.com/mt-sre/devkube/devcluster"
)

type Cluster struct {
	devcluster.Cluster
	KindName     string
	KindProvider KindProvider
}

func (c Cluster) ListNodes() ([]nodes.Node, error) {
	return c.KindProvider.ListNodes(c.KindName)
}

func (c Cluster) SendImageArchiveToNodes(archive []byte) error {
	nodesList, err := c.KindProvider.ListNodes(c.KindName)
	if err != nil {
		return fmt.Errorf("failed to list the nodes of the KinD cluster: %w", err)
	}

	for _, node := range nodesList {
		if err := nodeutils.LoadImageArchive(node, bytes.NewReader(archive)); err != nil {
			return err
		}
	}
	return nil
}

func (c Cluster) Kubeconfig(internal bool) (string, error) {
	return c.KindProvider.KubeConfig(c.KindName, internal)
}
