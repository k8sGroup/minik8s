package podManager

import (
	"minik8s/cmd/kubelet/app/dockerClient"
	"minik8s/cmd/kubelet/app/message"
	"minik8s/cmd/kubelet/app/module"
	"minik8s/cmd/kubelet/app/pod"
	"unsafe"
)

type PodManager struct {
	pods []*pod.Pod
}

var instance *PodManager

func GetPodManager() *PodManager {
	if instance == nil {
		instance = new(PodManager)
		return instance
	} else {
		return instance
	}
}

func (p PodManager) AddPodFromConfig(config module.Config) (*pod.Pod, error) {
	//form containers
	newPod := &pod.Pod{}
	newPod.Name = config.MetaData.Name
	//newPod.LabelMap["app"] = config.MetaData.Labels.App
	commandWithConfig := &message.CommandWithConfig{}
	commandWithConfig.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	commandWithConfig.Group = config.Spec.Containers
	response := dockerClient.HandleCommand(&(commandWithConfig.Command))
	if response.Err != nil {
		return nil, response.Err
	}
	responseWithContainersId := (*message.ResponseWithContainIds)(unsafe.Pointer(response))
	newPod.ContainerIds = responseWithContainersId.ContainersIds
	p.pods = append(p.pods, newPod)
	return newPod, nil
}

func (p PodManager) PullImages(images []string) error {
	commandWithImages := &message.CommandWithImages{}
	commandWithImages.CommandType = message.COMMAND_PULL_IMAGES
	commandWithImages.Images = images
	response := dockerClient.HandleCommand(&(commandWithImages.Command))
	return response.Err
}
