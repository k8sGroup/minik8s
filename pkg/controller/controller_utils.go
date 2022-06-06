package controller

import (
	"minik8s/object"
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
	return (object.PodExit != p.Status.Phase) && (object.Failed != p.Status.Phase)
}
