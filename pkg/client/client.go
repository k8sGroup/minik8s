package client

import (
	"context"
	"minik8s/object"
	"net/http"
	"net/url"
)

type Config struct {
	Host string
}

//type Interface interface {
//	Verb(verb string) *Request
//	Post() *Request
//	Put() *Request
//	Get() *Request
//	Delete() *Request
//}

type RESTClient struct {
	base   *url.URL
	Client *http.Client
}

func (r RESTClient) CreatePods(ctx context.Context, template *object.PodTemplateSpec) error {
	//pod, _ := GetPodFromTemplate(template)
	//newPod, _ := r.Create(ctx, pod)
	//klog.Infof("[RealPodControl] CreatePods %+v", newPod)
	return nil
}

func (r RESTClient) DeletePod(ctx context.Context, podID string) error {
	//err := r.Pods(namespace).Delete(ctx, podID)
	return nil
}

// GetPodFromTemplate TODO: type conversion
func GetPodFromTemplate(template *object.PodTemplateSpec) (*object.Pod, error) {
	return nil, nil
}
