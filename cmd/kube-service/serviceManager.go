package main

import (
	"minik8s/pkg/client"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/service"
)

func main() {
	clientConfig := client.Config{Host: "127.0.0.1:8080"}
	service.NewManager(listerwatcher.DefaultConfig(), clientConfig)
	select {}
}
