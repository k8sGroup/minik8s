package main

import (
	"bufio"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/netSupport/netconfig"
	"os"
	"strconv"
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
func formServerConfig(trans *object.DnsAndTrans) []string {
	var result []string
	result = append(result, "    server {", "        listen 80 ;")
	result = append(result, fmt.Sprintf("        server_name %s;", trans.Spec.Host))
	for _, val := range trans.Spec.Paths {
		result = append(result, fmt.Sprintf("        location ~ %s {", val.Name))
		result = append(result, fmt.Sprintf("            proxy_pass http://%s:%s;", val.Ip, val.Port))
		result = append(result, "        }")
	}
	result = append(result, "       }")
	return result
}
func writeNginxConfig(trans *object.DnsAndTrans) {
	var content []string
	content = append(content, "error_log stderr;")
	content = append(content, "events { worker_connections  1024; }")
	content = append(content, "http {", "    access_log /dev/stdout combined;")
	content = append(content, formServerConfig(trans)...)
	content = append(content, "}")
	//test
	fmt.Println(content)
	f, err := os.OpenFile(netconfig.NginxPathPrefix+"/"+trans.MetaData.Name+"/"+netconfig.NginxConfigFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("[dnsConfigWriter] writerNginxConfig error" + err.Error())
		return
	}
	w := bufio.NewWriter(f)
	for _, v := range content {
		fmt.Fprintln(w, v)
	}
	err = w.Flush()
	if err != nil {
		fmt.Println("[dnsConfigWriter] writerNginxConfig error" + err.Error())
		return
	}
	return
}
func getCpu(input string) int64 {
	len := len(input)
	result := 0.0
	if input[len-1] == 'm' {
		result, _ = strconv.ParseFloat(input[:len-1], 32)
		result *= 1e6
	} else {
		result, _ = strconv.ParseFloat(input[:len], 32)
		result *= 1e9
	}
	return int64(result)
}
func getMemory(input string) int64 {
	len := len(input)
	result, _ := strconv.Atoi(input[:len-1])
	mark := input[len-1]
	if mark == 'K' || mark == 'k' {
		result *= 1024
	} else if mark == 'M' || mark == 'm' {
		result *= 1024 * 1024
	} else {
		//G
		result *= 1024 * 1024 * 1024
	}
	return int64(result)
}
func main() {
	//fmt.Println(getCpu("0.5"))
	//fmt.Println(getCpu("1.8"))
	//fmt.Println(getCpu("65m"))
	//fmt.Println(getCpu("3.4m"))
	fmt.Println(getMemory("10k"))
	fmt.Println(getMemory("10G"))
}
