package main

import (
	"fmt"
	"minik8s/pkg/client"
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
	config := &client.Config{Host: "127.17.0.1" + ":8080"}
	restClient := client.RESTClient{
		Base: "http://" + config.Host,
	}
	attachUrl := "/registry/pod/default/" + "1111111"
	resp, err := client.Get(restClient.Base + attachUrl)
	if err != nil {
		fmt.Println(err)
	}
	if resp != nil {
		fmt.Println(resp)
	}
}
