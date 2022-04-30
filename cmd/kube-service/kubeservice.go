package main

import (
	"context"
	"fmt"
	"minik8s/pkg/kubelet/dockerClient"
)

func main() {
	cli, _ := dockerClient.GetNewClient()
	res, _ := cli.ContainerInspect(context.Background(), "pause_example_nginx_example_ghost")
	fmt.Println(res)

}
