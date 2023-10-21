// Package devkind manages kind clusters.
package devkind

import (
	"fmt"
	"os"
	"slices"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kindconfigv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"

	"github.com/mt-sre/devkube/devcluster"
)

type KindProvider interface {
	ListNodes(name string) ([]nodes.Node, error)
	KubeConfig(name string, internal bool) (string, error)
	Delete(name string, explicitKubeConfigPath string) error
	List() ([]string, error)
	Create(name string, opts ...kindcluster.CreateOption) error
}

type Kind struct{ Provider KindProvider }

func (k Kind) GetKindCluster(name string) (Cluster, error) {
	kubeconfig, err := k.Provider.KubeConfig(name, false)
	if err != nil {
		return Cluster{}, err
	}

	cc, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfig))
	if err != nil {
		return Cluster{}, err
	}

	rcc, err := cc.ClientConfig()
	if err != nil {
		return Cluster{}, err
	}

	cli, err := client.New(rcc, client.Options{})
	if err != nil {
		return Cluster{}, err
	}

	return Cluster{Cluster: devcluster.Cluster{Cli: cli}, KindName: name, KindProvider: k.Provider}, nil
}

func (k Kind) DeleteKindClusterByName(name string, kubeconfigPath string) error {
	return k.Provider.Delete(name, kubeconfigPath)
}

func (k Kind) ListKindClusterNames() ([]string, error) { return k.Provider.List() }

func (k Kind) CreateKindCluster(name, kubeconfigPath string, kindCfg kindconfigv1alpha4.Cluster) (Cluster, error) {
	kindCfg.TypeMeta = kindconfigv1alpha4.TypeMeta{Kind: "Cluster", APIVersion: "kind.x-k8s.io/v1alpha4"}
	kindconfigv1alpha4.SetDefaultsCluster(&kindCfg)

	// TODO why is this here?
	if _, err := os.Lstat("/dev/dm-0"); err == nil {
		for i := range kindCfg.Nodes {
			kindCfg.Nodes[i].ExtraMounts = append(kindCfg.Nodes[i].ExtraMounts,
				kindconfigv1alpha4.Mount{
					HostPath:      "/dev/dm-0",
					ContainerPath: "/dev/dm-0",
					Propagation:   kindconfigv1alpha4.MountPropagationHostToContainer,
				},
			)
		}
	}

	kindOpts := []kindcluster.CreateOption{
		kindcluster.CreateWithV1Alpha4Config(&kindCfg),
		kindcluster.CreateWithKubeconfigPath(kubeconfigPath),
		kindcluster.CreateWithDisplayUsage(false),
		kindcluster.CreateWithDisplaySalutation(false),
		kindcluster.CreateWithWaitForReady(5 * time.Minute),
		kindcluster.CreateWithRetain(false),
	}
	if err := k.Provider.Create(name, kindOpts...); err != nil {
		return Cluster{}, fmt.Errorf("failed to create the cluster: %w", err)
	}

	return k.GetKindCluster(name)
}

func (k Kind) CreateOrGetKindCluster(name, kubeconfigPath string, kindCfg kindconfigv1alpha4.Cluster) (Cluster, error) {
	names, err := k.ListKindClusterNames()
	if err != nil {
		return Cluster{}, fmt.Errorf("list kind clusters: %w", err)
	}
	if !slices.Contains(names, name) {
		return k.CreateKindCluster(name, kubeconfigPath, kindCfg)
	}
	return k.GetKindCluster(name)
}

func (k Kind) CreateOrRecreateKindCluster(
	name, kubeconfigPath string, kindCfg kindconfigv1alpha4.Cluster,
) (Cluster, error) {
	names, err := k.ListKindClusterNames()
	if err != nil {
		return Cluster{}, fmt.Errorf("list kind clusters: %w", err)
	}

	if !slices.Contains(names, name) {
		if err := k.DeleteKindClusterByName(name, kubeconfigPath); err != nil {
			return Cluster{}, fmt.Errorf("delete kind cluster: %w", err)
		}
	}

	return k.CreateKindCluster(name, kubeconfigPath, kindCfg)
}
