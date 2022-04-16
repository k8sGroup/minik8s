package typed

import (
	"context"
	"encoding/json"
	"minik8s/k8s.io/client/rest"
	"minik8s/object"
)

type PodsGetter interface {
	Pods(namespace string) PodInterface
}

type PodInterface interface {
	Create(ctx context.Context, pod *object.Pod) (*object.Pod, error)
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*object.Pod, error)
	List(ctx context.Context, opts object.ListOptions) (*object.PodList, error)
}

type pods struct {
	client rest.Interface
	ns     string
}

func NewPods(c rest.Interface, namespace string) *pods {
	return &pods{
		client: c,
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

func (c *pods) Delete(ctx context.Context, name string) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("pods").
		Name(name).
		Do(ctx).
		Error()
}

func (c *pods) Get(ctx context.Context, name string) (result *object.Pod, err error) {
	result = &object.Pod{}
	res := c.client.Get().
		Namespace(c.ns).
		Resource("pods").
		Name(name).
		Do(ctx)
	body := res.GetBody()
	err = json.Unmarshal(body, result)
	return result, err
}

func (c *pods) List(ctx context.Context, opts object.ListOptions) (result *object.PodList, err error) {
	result = &object.PodList{}
	res := c.client.Get().
		Namespace(c.ns).
		Resource("pods").
		Do(ctx)
	body := res.GetBody()
	err = json.Unmarshal(body, result)
	return
}
