package config

import (
	"minik8s/cmd/kubelet/app/pod"
	"minik8s/cmd/kubelet/app/podManager"
	"minik8s/cmd/kubelet/app/types"
)

type PodConfig struct {
	//podManager 与 dockerClient交互
	podManager *podManager.PodManager
	//建立映射  source(pod源)--(map<pod id - *pod>)
	pods map[string]map[string]*pod.Pod

	//管道
	updates chan types.PodUpdate
}
