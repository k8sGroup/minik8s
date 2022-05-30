package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"minik8s/pkg/kubelet/dockerClient"
)

func main() {
	cli, _ := dockerClient.GetNewClient()
	res, _ := dockerClient.GetAllContainers()
	for _, v := range res {
		cli.ContainerRemove(context.Background(), v.ID, types.ContainerRemoveOptions{})
	}
}
