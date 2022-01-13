package dev

import (
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
)

type WithLogger logr.Logger

func (l WithLogger) ApplyToEnvironmentConfig(c *EnvironmentConfig) {
	c.Logger = logr.Logger(l)
}

func (l WithLogger) ApplyToClusterConfig(c *ClusterConfig) {
	c.Logger = logr.Logger(l)
}

func (l WithLogger) ApplyToWaiterConfig(c *WaiterConfig) {
	c.Logger = logr.Logger(l)
}

type WithInterval time.Duration

func (i WithInterval) ApplyToWaiterConfig(c *WaiterConfig) {
	c.Interval = time.Duration(i)
}

type WithTimeout time.Duration

func (t WithTimeout) ApplyToWaiterConfig(c *WaiterConfig) {
	c.Timeout = time.Duration(t)
}

type WithSchemeBuilder runtime.SchemeBuilder

func (sb WithSchemeBuilder) ApplyToClusterConfig(c *ClusterConfig) {
	c.SchemeBuilder = runtime.SchemeBuilder(sb)
}

type WithNewWaiterFunc ClusterNewWaiterFunc

func (f WithNewWaiterFunc) ApplyToClusterConfig(c *ClusterConfig) {
	c.NewWaiter = ClusterNewWaiterFunc(f)
}

type WithWaitOptions []WaitOption

func (opts WithWaitOptions) ApplyToClusterConfig(c *ClusterConfig) {
	c.WaitOptions = []WaitOption(opts)
}
