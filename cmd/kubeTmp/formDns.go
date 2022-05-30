package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/client"
)

func main() {
	clientConfig := client.Config{Host: "127.0.0.1" + ":8080"}
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	data, err := ioutil.ReadFile("/home/minik8s/test/dnsAndTrans/dnsTest.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	dnsAndTrans := &object.DnsAndTrans{}
	err = yaml.Unmarshal([]byte(data), dnsAndTrans)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dnsAndTrans)
	err = restClient.UpdateDnsAndTrans(dnsAndTrans)
	if err != nil {
		fmt.Println(err)
	}
}
