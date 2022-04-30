package main

import (
	"fmt"
	"strings"
)

func main() {
	//dockerIp := "172.17.0.2"
	//dockerPort := "80"
	//hostPort := "8080"
	//err := iptablesManager.AddDockerChainMappingRule(dockerIp, dockerPort, hostPort)
	//if err != nil {
	//	fmt.Printf(err.Error())
	//}
	//buffer, err := iptablesManager.GetNatDockerChainRules()
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	//fmt.Println(buffer)
	//	for _, value := range buffer {
	//		fmt.Println(value)
	//	}
	//}
	ss := "2        2   104 DNAT       tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:8080 to:172.17.0.2:80"
	index := strings.Index(ss, "dpt:")
	index += 4
	end := index + 1
	for {
		if ss[end] > '9' || ss[end] < '0' {
			break
		}
		end++
	}
	hostPort := ss[index:end]
	index = strings.Index(ss, "to:")
	index += 3
	ipAndPort := ss[index:]
	index = strings.Index(ipAndPort, ":")
	ip := ipAndPort[0:index]
	port := ipAndPort[index+1:]
	//end = index

	//index = strings.Index(ipAndPort, ":")
	//ip := ss[end : index-1]
	//port := ss[index+1:]
	fmt.Println(hostPort)
	fmt.Println(ip)
	fmt.Println(port)
}
