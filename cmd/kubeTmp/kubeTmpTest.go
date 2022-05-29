package main

import (
	"fmt"
	"minik8s/pkg/etcdstore/serviceConfigStore"
	"time"
)

func Test() {
	timer := time.NewTicker(3 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				fmt.Println("111111")
			}
		}
	}()

}
func main() {
	var name string
	var ip string
	for {
		fmt.Println("输入名字")
		fmt.Scanln(&name)
		//fmt.Println("输入ip")
		//fmt.Scanln(&ip)
		ok, alloc := serviceConfigStore.JudgeAndAllocClusterIp(name, ip)
		fmt.Println(ok)
		fmt.Println(alloc)
	}
	//writer := kubeproxy.NewDnsConfigWriter(nil)
	//input := &object.DnsAndTrans{
	//	MetaData: object.ObjectMeta{
	//		Name: "test",
	//	},
	//	Spec: object.DnsAndTransSpec{
	//		Host: "aaa.bbb.ccc",
	//		Paths: []object.TransPath{
	//			object.TransPath{
	//				Name:    "/hong",
	//				Service: "hong",
	//				Ip:      "11.11.11.11",
	//				Port:    "80",
	//			},
	//		},
	//	},
	//}
	//writer.AddDnsAndTrans("test", input)
	//input.Spec.Paths[0].Ip = "10.10.10.22"
	//input.Spec.Paths[0].Name = "/ssss"
	//input.Spec.Paths[0].Port = "100"
	//input.Spec.Host = "aasd.fff"
	//writer.AddDnsAndTrans("oooo", input)
}
