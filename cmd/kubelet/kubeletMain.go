package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
	"os"
)

//var (
//	LOCAL   = "127.0.0.1"
//	REMOTE  = "192.168.1.7"
//	MASTER  = "192.168.1.4"
//	MASTER2 = "10.119.11.164"
//)

func parseConfigFile(path string) *object.Node {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("ReadFile in %s fail, use default config", os.Args[1])
		return nil
	}
	node := &object.Node{}
	err = yaml.Unmarshal([]byte(data), node)
	if err != nil {
		fmt.Printf("file in %s unmarshal fail, use default config", path)
		return nil
	}
	return node
}

func main() {
	var node *object.Node
	masterIp := netconfig.MasterIp
	if len(os.Args) != 1 {
		//参数应该为yaml文件路径,进行解析
		node = parseConfigFile(os.Args[1])
		if node != nil {
			masterIp = node.MasterIp
		}
	}
	clientConfig := client.Config{Host: masterIp + ":8080"}
	kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(masterIp), clientConfig, node)
	kube.Run()
	fmt.Printf("kube run emd...\n")
	select {}
	//var m int
	//for {
	//	fmt.Println("查看错误信息\n")
	//	fmt.Scanln(&m)
	//	fmt.Println(kube.Err)
	//}
}
