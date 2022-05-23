package main

import "minik8s/cmd/kubectl/app"

var (
	LOCAL  = "127.0.0.1"
	REMOTE = "10.119.11.164"
)

func main() {
	//data, err := ioutil.ReadFile("/home/minik8s/test/pod/example.yaml")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//pod := &object.Pod{}
	//err = yaml.Unmarshal([]byte(data), &pod)
	//fmt.Println(*pod)
	//clientConfig := client.Config{Host: LOCAL + ":8080"}
	//restClient := client.RESTClient{
	//	Base: "http://" + clientConfig.Host,
	//}
	//pod.UID = uuid.NewV4().String()
	//restClient.UpdateConfigPod(pod)
	//podName := pod.Name
	//var m int
	//for {
	//	fmt.Println("输入操作类型\n")
	//	fmt.Println("1. 查看配置pod\n 2. 查看运行时pod \n 3.删除pod\n 4.修改pod部署节点。\n")
	//	fmt.Scanln(&m)
	//	switch m {
	//	case 1:
	//		value, err2 := restClient.GetConfigPod(podName)
	//		if err2 != nil {
	//			fmt.Println(err2)
	//		} else {
	//			fmt.Println(value)
	//		}
	//		continue
	//	case 2:
	//		value, err2 := restClient.GetRuntimePod(podName)
	//		if err2 != nil {
	//			fmt.Println(err2)
	//		} else {
	//			fmt.Println(value)
	//		}
	//		continue
	//	case 3:
	//		value, err2 := restClient.GetConfigPod(podName)
	//		if err2 != nil {
	//			fmt.Println(err2)
	//		} else {
	//			value.Status.Phase = object.PodDelete
	//			err2 = restClient.UpdateConfigPod(value)
	//			if err2 != nil {
	//				fmt.Println(err2)
	//			}
	//		}
	//		continue
	//	case 4:
	//		value, err2 := restClient.GetConfigPod(podName)
	//		if err2 != nil {
	//			fmt.Println(err2)
	//		} else {
	//			if value.Spec.NodeName == "node1" {
	//				value.Spec.NodeName = "node2"
	//			} else {
	//				value.Spec.NodeName = "node1"
	//			}
	//			err2 = restClient.UpdateConfigPod(value)
	//			if err2 != nil {
	//				fmt.Println(err2)
	//			}
	//		}
	//		continue
	//	}
	//}
	_ = app.Execute()
}
