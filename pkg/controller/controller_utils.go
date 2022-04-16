package controller

import (
	"context"
	"minik8s/k8s.io/client/typed"
	"minik8s/object"
	"minik8s/pkg/klog"
)

type PodControlInterface interface {
	CreatePods(ctx context.Context, namespace string, template *object.PodTemplateSpec) error
	DeletePod(ctx context.Context, namespace string, podID string) error
	AutoUpdatePodsInfo() error
}

type RealPodControl struct {
	//RESTClient() rest.Interface
	typed.PodsGetter
}

//func (r RealPodControl) Pods(namespace string) typed.PodInterface {
//	return typed.NewPods(r, namespace)
//}
//
//func (r RealPodControl) RESTClient() rest.Interface {
//	return r.restClient
//}

func (r RealPodControl) CreatePods(ctx context.Context, namespace string, template *object.PodTemplateSpec) error {
	pod, _ := GetPodFromTemplate(template)
	newPod, _ := r.Pods(namespace).Create(ctx, pod)
	klog.Infof("[RealPodControl] CreatePods %+v", newPod)
	return nil
}

func (r RealPodControl) DeletePod(ctx context.Context, namespace string, podID string) error {
	err := r.Pods(namespace).Delete(ctx, podID)
	return err
}

// GetPodFromTemplate TODO: type conversion
func GetPodFromTemplate(template *object.PodTemplateSpec) (*object.Pod, error) {
	return nil, nil
}
