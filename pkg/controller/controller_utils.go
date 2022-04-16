package controller

import (
	"context"
	"minik8s/k8s.io/client/core"
	"minik8s/object"
)

type PodControlInterface interface {
	CreatePods(ctx context.Context, namespace string, template *object.PodTemplateSpec) error
}

type RealPodControl struct {
	KubeClient core.CoreInterface
}

func (r RealPodControl) CreatePods(ctx context.Context, namespace string, template *object.PodTemplateSpec) error {
	pod, _ := GetPodFromTemplate(template)
	newPod, _ := r.KubeClient.Pods(namespace).Create(ctx, pod)
	return nil
}

func GetPodFromTemplate(template *object.PodTemplateSpec) (*object.Pod, error) {
	return nil, nil
}
