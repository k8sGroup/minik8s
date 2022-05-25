package main

import (
	"minik8s/pkg/client"
	"minik8s/pkg/kubeproxy"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
)

func main() {
	clientConfig := client.Config{Host: netconfig.MasterIp + ":8080"}
	kubeProxy := kubeproxy.NewKubeProxy(listerwatcher.GetLsConfig(netconfig.MasterIp), clientConfig)
	kubeProxy.StartKubeProxy()
	select {}
}
