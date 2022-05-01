package main

import (
	"fmt"
	"minik8s/pkg/kubeproxy/ipManager"
)

func main() {
	//ip, err := tools.GetIPv4ByInterface("ens33")
	//fmt.Println(ip)
	//fmt.Println(err)
	res, err := ipManager.GetRouteInfo()
	fmt.Println(res)
	fmt.Println(err)
}
