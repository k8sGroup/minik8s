package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/client"
)

func main() {
	data, err := ioutil.ReadFile("/home/minik8s/build/buildPod/coreDnsPod.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	pod := &object.Pod{}
	err = yaml.Unmarshal([]byte(data), pod)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(*pod)
	}
	clientConfig := client.Config{Host: "127.0.0.1" + ":8080"}
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	restClient.UpdateConfigPod(pod)
	data, err = ioutil.ReadFile("/home/minik8s/build/buildPod/gateWayPod.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	pod = &object.Pod{}
	err = yaml.Unmarshal([]byte(data), pod)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*pod)
	restClient.UpdateConfigPod(pod)
}
