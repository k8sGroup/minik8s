package main

import (
	"fmt"
	"minik8s/pkg/service"
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
	res := service.GetCoreDnsServiceModule()
	fmt.Println(res.Spec.PodNameAndIps)
	fmt.Println(res)
	fmt.Println("-------------------------------------------\n")
	res2 := service.GetCoreDnsRsModule()
	fmt.Println(res2)
	fmt.Println("-------------------------------------------\n")
	res3 := service.GetGateWayRsModule("test")
	fmt.Println(res3)
	fmt.Println("-------------------------------------------\n")
	res4 := service.GetGateWayServiceModule("test")
	fmt.Println(res4)
	fmt.Println("-------------------------------------------\n")
}
