package kubeproxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/kubelet/dockerClient"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
	"os"
	"os/exec"
	"strings"
	"time"
)

//更新本地的coreDns和nginx的配置文件

type DnsConfigWriter struct {
	key2DnsAnsTrans map[string]*object.DnsAndTrans
	ls              *listerwatcher.ListerWatcher
	Client          client.RESTClient
	stopChannel     <-chan struct{}
}

func NewDnsConfigWriter(lsConfig *listerwatcher.Config, clientConfig client.Config) *DnsConfigWriter {
	res := &DnsConfigWriter{}
	res.key2DnsAnsTrans = make(map[string]*object.DnsAndTrans)
	res.stopChannel = make(chan struct{})
	res.Client = client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[DnsConfigWriter] newDnsConfigWriter Error")
		fmt.Println(err)
	}
	res.ls = ls
	res.register()
	return res
}

func (d *DnsConfigWriter) register() {
	watchFunc := func() {
		for {
			err := d.ls.Watch(config.DnsAndTransPrefix, d.watchDnsAndTrans, d.stopChannel)
			if err != nil {
				fmt.Println("[dnsConfigWriter] watch error" + err.Error())
				time.Sleep(10 * time.Second)
			} else {
				return
			}
		}
	}
	go watchFunc()
}

//func (d *DnsConfigWriter) update(DnsAndTransName string) {
//	d.writeCoreDnsConfig()
//	d.writeNginxConfig()
//	reloadNginxConf()
//}
func (d *DnsConfigWriter) formDir(DnsAndTransName string) {
	args := fmt.Sprintf("-r %s %s", netconfig.NginxDirModelPath, netconfig.NginxPathPrefix+"/"+DnsAndTransName)
	res, err := execCmdWithOutput("cp", args)
	if err != nil {
		fmt.Println("[dnsConfigWriter]formDir fail")
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
}
func (d *DnsConfigWriter) deleteDir(DnsAndTransName string) {
	args := fmt.Sprintf("-rf %s", netconfig.NginxPathPrefix+"/"+DnsAndTransName)
	res, err := execCmdWithOutput("rm", args)
	if err != nil {
		fmt.Println("[dnsConfigWriter]formDir fail")
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
}
func (d *DnsConfigWriter) watchDnsAndTrans(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		//实际的删除通过设置status进行
		return
	} else {
		DnsAndTrans := &object.DnsAndTrans{}
		err := json.Unmarshal(res.ValueBytes, DnsAndTrans)
		if err != nil {
			fmt.Println("[dnsConfigWriter] watchDnsAndTrans error" + err.Error())
			return
		}
		switch DnsAndTrans.Status.Phase {
		case "":
			//初次加入，需要生成文件夹
			d.key2DnsAnsTrans[res.Key] = DnsAndTrans
			//生成文件
			d.formDir(DnsAndTrans.MetaData.Name)
			//写nginx的配置
			d.writeNginxConfig(DnsAndTrans)
			DnsAndTrans.Status.Phase = object.FileCreated
			d.Client.UpdateDnsAndTrans(DnsAndTrans)
			break
		case object.FileCreated:
			//自己提交的应该是， 不用管
			break
		case object.ServiceCreated:
			//服务已经部署, 可能是用户端的更新或者是gateWay服务的生成
			d.key2DnsAnsTrans[res.Key] = DnsAndTrans
			d.writeNginxConfig(DnsAndTrans)
			reloadNginxConf(DnsAndTrans.MetaData.Name)
			d.writeCoreDnsConfig()
			break
		case object.Delete:
			_, ok := d.key2DnsAnsTrans[res.Key]
			if !ok {
				return
			} else {
				delete(d.key2DnsAnsTrans, res.Key)
				d.writeCoreDnsConfig()
				//删除nginx文件夹
				d.deleteDir(DnsAndTrans.MetaData.Name)
			}
			break
		}
	}
}
func (d *DnsConfigWriter) writeCoreDnsConfig() {
	//test
	f, err := os.OpenFile(netconfig.CoreDnsConfigPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println("[dnsConfigWriter] writeCoreDnsConfig fail")
		return
	}
	w := bufio.NewWriter(f)
	for _, v := range d.key2DnsAnsTrans {
		if v.Status.Phase != object.ServiceCreated {
			continue
		}
		//加入网关ip - name的条目
		lineStr := fmt.Sprintf("%s %s", v.Spec.GateWayIp, v.Spec.Host)
		fmt.Fprintln(w, lineStr)
	}
	err = w.Flush()
	if err != nil {
		fmt.Println("[dnsConfigWriter] writeCoreDnsConfig fail " + err.Error())
	}
	return
}

func (d *DnsConfigWriter) writeNginxConfig(trans *object.DnsAndTrans) {
	var content []string
	content = append(content, "error_log stderr;")
	content = append(content, "events { worker_connections  1024; }")
	content = append(content, "http {", "    access_log /dev/stdout combined;")
	content = append(content, d.formServerConfig(trans)...)
	content = append(content, "}")
	//test
	f, err := os.OpenFile(netconfig.NginxPathPrefix+trans.MetaData.Name+"/"+netconfig.NginxConfigFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
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
func reloadNginxConf(DnsAndTransName string) {
	//先获取是否有对应的容器在运行
	res, err := dockerClient.GetRunningContainers()
	if err != nil {
		fmt.Println("[dnsConfigWriter] reloadNginx error" + err.Error())
	}
	var containerIds []string
	for _, val := range res {
		if strings.Contains(val.Names[0], netconfig.GateWayContainerPrefix+DnsAndTransName) {
			containerIds = append(containerIds, val.ID)
		}
	}
	for _, containerId := range containerIds {
		args := fmt.Sprintf("exec %s nginx -s reload", containerId)
		res, err := execCmdWithOutput("docker", args)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(res)
		}
	}
}
func execCmdWithOutput(exc string, args string) ([]string, error) {
	cmd := exec.Command(exc, strings.Split(args, " ")...)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	//创建一个流来读取管道内容，一行一行读取
	reader := bufio.NewReader(stdout)
	var result []string
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		result = append(result, line)
	}
	err = cmd.Wait()
	return result, err
}
func (d *DnsConfigWriter) formServerConfig(trans *object.DnsAndTrans) []string {
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
