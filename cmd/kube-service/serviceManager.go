package main

import (
	"minik8s/pkg/client"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/service"
)

// Deprecated:  This module has already been merged into controller manager.
func main() {
	clientConfig := client.Config{Host: "127.0.0.1:8080"}
	service.NewManager(listerwatcher.DefaultConfig(), clientConfig)
	select {}
}
