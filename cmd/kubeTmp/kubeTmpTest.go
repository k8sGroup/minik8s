package main

import (
	"bufio"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/netSupport/netconfig"
	"os"
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
func main() {
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
	dnsAndTrans.Spec.Paths[0].Ip = "10.10.10.1"
	writeNginxConfig(dnsAndTrans)
}
