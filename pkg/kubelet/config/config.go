package config

import (
	"minik8s/object"
	"minik8s/pkg/kubelet/types"
	"sync"
)

type PodConfig struct {
	// TODO: two pod manager?
	//podManager 与 dockerClient交互
	//podManager *podManager.PodManager
	podLock sync.RWMutex
	//建立映射  source(pod源)--(map<pod id - *pod>)
	pods map[string]map[string]*object.Pod
	//管道
	updates chan types.PodUpdate
}

// NewPodConfig TODO: complete new pod configuration
func NewPodConfig() *PodConfig {
	updates := make(chan types.PodUpdate, 50)
	podConfig := &PodConfig{
		updates: updates,
	}
	return podConfig
}

func (pc PodConfig) GetUpdates() chan types.PodUpdate {
	return pc.updates
}
