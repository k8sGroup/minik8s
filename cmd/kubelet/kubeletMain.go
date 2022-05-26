package main

import (
	"fmt"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
)

//var (
//	LOCAL   = "127.0.0.1"
//	REMOTE  = "192.168.1.7"
//	MASTER  = "192.168.1.4"
//	MASTER2 = "10.119.11.164"
//)

func main() {
	// host is the address of master node
	clientConfig := client.Config{Host: netconfig.MasterIp + ":8080"}
	//kube := kubelet.NewKubelet(listerwatcher.GetLsConfig("192.168.1.7"), clientConfig)
	kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(netconfig.MasterIp), clientConfig)
	kube.Run()
	fmt.Printf("kube run emd...\n")
	var m int
	for {
		fmt.Println("查看错误信息\n")
		fmt.Scanln(&m)
		fmt.Println(kube.Err)
	}
}
