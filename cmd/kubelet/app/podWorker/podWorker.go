package podWorker

import (
	"minik8s/cmd/kubelet/app/dockerClient"
	"minik8s/cmd/kubelet/app/message"
)

//用于Pod 与docker client之间的交互
type PodWorker struct {
}

func (podWorker PodWorker) syncLoop(commands <-chan message.PodCommand, responses chan<- message.PodResponse) {
	for {
		select {
		case command, _ := <-commands:
			res := dockerClient.HandleCommand(command.ContainerCommand)
			result := message.PodResponse{
				ContainerResponse: res,
				PodUid:            command.PodUid,
				PodResponseType:   command.PodCommandType,
			}
			responses <- result
		}
	}
}
