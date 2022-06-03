package main

import (
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/mesh"
	"minik8s/pkg/netSupport/netconfig"
	"syscall"
)

func main() {
	syscall.Setuid(1337)
	p := mesh.NewProxy(listerwatcher.GetLsConfig(netconfig.MasterIp))
	addresses := []mesh.Address{{mesh.InIP, mesh.InPort, mesh.DirIn}, {mesh.OutIP, mesh.OutPort, mesh.DirOut}}
	//addresses := []mesh.Address{{mesh.InIP, mesh.InPort, mesh.DirIn}}
	p.Init(addresses)
	select {}
}
