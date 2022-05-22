package controller

import (
	"minik8s/object"
	"minik8s/pkg/kubelet/pod"
)

func FilterActivePods(pods []*object.Pod) []*object.Pod {
	var result []*object.Pod
	for _, p := range pods {
		if IsPodActive(p) {
			result = append(result, p)
		}
	}
	return result
}

func IsPodActive(p *object.Pod) bool {
	return (pod.POD_EXITED_STATUS != p.Status.Phase) && (pod.POD_FAILED_STATUS != p.Status.Phase)
}
