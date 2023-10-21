//nolint:revive
package devkind

import (
	"github.com/stretchr/testify/mock"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
)

type KindProviderMock struct{ mock.Mock }

func (k *KindProviderMock) ListNodes(name string) ([]nodes.Node, error) {
	args := k.Called(name)
	return args.Get(0).([]nodes.Node), args.Error(1)
}

func (k *KindProviderMock) KubeConfig(name string, internal bool) (string, error) {
	args := k.Called(name, internal)
	return args.String(0), args.Error(1)
}

func (k *KindProviderMock) Delete(name string, explicitKubeConfigPath string) error {
	return k.Called(name, explicitKubeConfigPath).Error(0)
}

func (k *KindProviderMock) List() ([]string, error) {
	args := k.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (k *KindProviderMock) Create(name string, opts ...kindcluster.CreateOption) error {
	return k.Called(name, opts).Error(0)
}
