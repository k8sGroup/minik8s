package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"minik8s/cmd/kubelet/app/module"
	"minik8s/cmd/kubelet/app/podManager"
)

func main() {

	p := podManager.NewPodManager()
	data, err := ioutil.ReadFile("D:\\goLandProject\\minik8s\\minik8s\\cmd\\kubelet\\example.yaml")
	if err != nil {
		fmt.Printf(err.Error())
	}
	conf := &module.Config{}
	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		fmt.Printf(err.Error())
	}
	p.AddPodFromConfig(*conf)
	for {
		fmt.Printf("输入操作类型:\n" +
			"1、新建pod\n" +
			"2、查询pod信息\n")
		var choice int
		fmt.Scanln(&choice)
		switch choice {
		case 1:
			fmt.Printf("输入pod配置文件路径\n")
			var path string
			fmt.Scanln(&path)
			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			conf := &module.Config{}
			err = yaml.Unmarshal([]byte(data), &conf)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			p.AddPodFromConfig(*conf)
			fmt.Println("pod配置已提交处理\n")
		case 2:
			fmt.Printf("输入pod名字\n")
			var name string
			fmt.Scanln(&name)
			resp, err := p.GetPodInfo(name)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			fmt.Printf(string(resp))
		}
	}
}
