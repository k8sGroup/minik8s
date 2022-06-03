package main

import (
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/mesh_sidecar"
	"minik8s/pkg/netSupport/netconfig"
)

func main() {
	sidecar := mesh_sidecar.NewSidecar(listerwatcher.GetLsConfig(netconfig.MasterIp))
	sidecar.Run()
	select {}
}
