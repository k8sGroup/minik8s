package typed

import (
	"context"
	"encoding/json"
	"minik8s/k8s.io/client/core"
	"minik8s/k8s.io/client/rest"
	"minik8s/object"
)

type PodsGetter interface {
	Pods(namespace string) PodInterface
}

type PodInterface interface {
	Create(ctx context.Context, pod *object.Pod) (*object.Pod, error)
}

type pods struct {
	client rest.Interface
	ns     string
}

func NewPods(c *core.CoreClient, namespace string) *pods {
	return &pods{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *pods) Create(ctx context.Context, pod *object.Pod) (result *object.Pod, err error) {
	result = &object.Pod{}
	res := c.client.Post().
		Namespace(c.ns).
		Resource("pods").
		Body(pod).
		Do(ctx)
	body := res.GetBody()
	err = json.Unmarshal(body, result)
	return result, nil
}
