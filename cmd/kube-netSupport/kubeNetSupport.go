package main

import (
	"fmt"
	"minik8s/pkg/client"
	"minik8s/pkg/kubeNetSupport"
	"minik8s/pkg/listerwatcher"
)

func main() {
	kubeNetSupport, err := kubeNetSupport.NewKubeNetSupport(listerwatcher.DefaultConfig(), client.DefaultClientConfig(), true)
	if err != nil {
		fmt.Println(err)
	}
	err = kubeNetSupport.StartKubeNetSupport()
	if err != nil {
		fmt.Println(err)
	}
	var m int
	for {
		fmt.Println("查看信息\n")
		fmt.Scanln(&m)
		fmt.Println(kubeNetSupport.GetKubeproxySnapShoot())
	}
}
