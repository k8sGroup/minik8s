package core

import (
	"minik8s/k8s.io/client/rest"
	"minik8s/k8s.io/client/typed"
)

type CoreInterface interface {
	RESTClient() rest.Interface

	typed.PodsGetter
}

type CoreClient struct {
	restClient rest.Interface
}

func (c *CoreClient) Pods(namespace string) typed.PodInterface {
	return typed.NewPods(c, namespace)
}

func (c *CoreClient) RESTClient() rest.Interface {
	return c.restClient
}
