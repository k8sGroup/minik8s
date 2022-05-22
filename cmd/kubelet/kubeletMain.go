package main

import (
	"fmt"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
)

//var (
//	LOCAL   = "127.0.0.1"
//	REMOTE  = "192.168.1.7"
//	MASTER  = "192.168.1.4"
//	MASTER2 = "10.119.11.164"
//)

func main() {
	// host is the address of master node
	clientConfig := client.Config{Host: netconfig.MasterIp + ":8080"}
	//kube := kubelet.NewKubelet(listerwatcher.GetLsConfig("192.168.1.7"), clientConfig)
	kube := kubelet.NewKubelet(listerwatcher.GetLsConfig(netconfig.MasterIp), clientConfig)
	kube.Run()
	//data, err := ioutil.ReadFile("./test/pod/example.yaml")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//pod := &object.Pod{}
	//err = yaml.Unmarshal([]byte(data), &pod)
	//err = kube.AddPod(pod)
	//if err != nil {
	//	fmt.Println(err)
	//}
	var m int
	for {
		fmt.Println("查看错误信息\n")
		fmt.Scanln(&m)
		fmt.Println(kube.Err)
	}
}

//
//func main() {
//	sysType := runtime.GOOS
//
//	// host is the address of master node
//	clientConfig := client.Config{Host: "127.0.0.1:8080"}
//	p := kubelet.NewKubelet(clientConfig)
//	err := p.Register()
//	var err error
//	var data []byte
//
//	if sysType == "linux" {
//		//InLinux
//		data, err = ioutil.ReadFile("/home/minik8s/cmd/kubelet/example.yaml")
//	}
//	if sysType == "windows" {
//		data, err = ioutil.ReadFile("D:\\goLandProject\\minik8s\\minik8s\\cmd\\kubelet\\exampleWin.yaml")
//	}
//
//	//InWindows
//
//	if err != nil {
//		fmt.Printf(err.Error())
//	}
//	conf := &module.Config{}
//	err = yaml.Unmarshal([]byte(data), &conf)
//	if err != nil {
//		fmt.Printf(err.Error())
//	}
//	p.AddPodFromConfig(*conf)
//	for {
//		fmt.Printf("输入操作类型:\n" +
//			"1、新建pod\n" +
//			"2、查询pod信息\n" +
//			"3、删除pod\n" +
//			"4、为pod添加端口映射\n" +
//			"5、查询端口映射信息\n" +
//			"6、删除端口映射\n")
//		var choice int
//		fmt.Scanln(&choice)
//		switch choice {
//		case 1:
//			fmt.Printf("输入pod配置文件路径\n")
//			var path string
//			fmt.Scanln(&path)
//			data, err := ioutil.ReadFile(path)
//			if err != nil {
//				fmt.Printf(err.Error())
//				continue
//			}
//			conf := &module.Config{}
//			err = yaml.Unmarshal([]byte(data), &conf)
//			if err != nil {
//				fmt.Printf(err.Error())
//				continue
//			}
//			p.AddPodFromConfig(*conf)
//			fmt.Println("pod配置已提交处理\n")
//		case 2:
//			fmt.Printf("输入pod名字\n")
//			var name string
//			fmt.Scanln(&name)
//			resp, err := p.GetPodInfo(name)
//			if err != nil {
//				fmt.Printf(err.Error())
//				continue
//			}
//			fmt.Printf(string(resp))
//		case 3:
//			fmt.Printf("输入pod名字\n")
//			var name string
//			fmt.Scanln(&name)
//			err := p.DeleteRuntimePod(name)
//			if err != nil {
//				fmt.Printf(err.Error())
//				continue
//			}
//			fmt.Printf("成功删除pod")
//		case 4:
//			fmt.Printf("输入选择的pod名\n")
//			var name string
//			fmt.Scanln(&name)
//			fmt.Printf("输入pod端口\n")
//			var podPort string
//			fmt.Scanln(&podPort)
//			var hostPort string
//			fmt.Printf("输入host端口\n")
//			fmt.Scanln(&hostPort)
//			_, err := p.AddPodPortMapping(name, podPort, hostPort)
//			if err != nil {
//				fmt.Printf(err.Error())
//			} else {
//				fmt.Printf("操作成功\n")
//			}
//		case 5:
//			fmt.Println(p.GetPodMappingInfo())
//		case 6:
//			fmt.Printf("输入选择的pod名\n")
//			var name string
//			fmt.Scanln(&name)
//			fmt.Printf("输入pod端口\n")
//			var podPort string
//			fmt.Scanln(&podPort)
//			var hostPort string
//			fmt.Printf("输入host端口\n")
//			fmt.Scanln(&hostPort)
//			err = p.RemovePortMapping(name, podPort, hostPort)
//			if err != nil {
//				fmt.Printf(err.Error())
//			} else {
//				fmt.Printf("操作成功")
//			}
//
//		}
//	}
//}
