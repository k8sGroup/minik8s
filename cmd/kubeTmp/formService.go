package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/client"
)

func main() {
	clientConfig := client.Config{Host: "127.0.0.1" + ":8080"}
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	data, err := ioutil.ReadFile("/home/minik8s/test/service/ghostService.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Print(data)
	service := &object.Service{}
	err = yaml.Unmarshal([]byte(data), service)
	fmt.Println(*service)
	err = restClient.UpdateService(service)
	if err != nil {
		fmt.Println(err)
		return
	}
}
