package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/client"
)

func main() {
	data, err := ioutil.ReadFile("/home/minik8s/test/pod/example.yaml")
	if err != nil {
		fmt.Println(err)
	}
	pod := &object.Pod{}
	err = yaml.Unmarshal([]byte(data), &pod)
	fmt.Println(*pod)
	clientConfig := client.Config{Host: "127.0.0.1" + ":8080"}
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	pod.Name = "example2222"
	pod.UID = uuid.NewV4().String()
	restClient.UpdateConfigPod(pod)
}
