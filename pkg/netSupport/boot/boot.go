package boot

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"minik8s/pkg/kubelet/dockerClient"
	"minik8s/pkg/netSupport/netconfig"
	"minik8s/pkg/netSupport/tools"
	"os"
	"os/exec"
	"strings"
	"time"
)

func BootFlannel() error {
	//先暂停所有的容器
	err := stopAllContainers()
	if err != nil {
		return err
	}
	//运行flannel插件
	go runFlanneld()
	time.Sleep(10 * time.Second)
	fmt.Println("run flannel finish")
	//运行DockerOpt
	err = runDockerOpt()
	if err != nil {
		return err
	}
	//修改配置文件
	err = ModifyDockerServiceConfig()
	if err != nil {
		return err
	}
	//重启docker 服务
	err = restartDockerService()
	if err != nil {
		return err
	}
	return nil
}

//暂停所有容器
func stopAllContainers() error {
	cli, err2 := dockerClient.GetNewClient()
	if err2 != nil {
		return err2
	}
	resp, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		return err
	}
	for _, value := range resp {
		err = cli.ContainerStop(context.Background(), value.ID, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func runFlanneld() error {
	args := fmt.Sprintf("--etcd-endpoints=%s --iface-regex=%s --ip-masq=true --etcd-prefix=%s --public-ip=%s",
		netconfig.EtcdEndPoint, netconfig.IFaceRegex, netconfig.EtcdNetworkPrefix, tools.GetDynamicIp())
	//cmd := exec.Command(netconfig.FlanneldPath, strings.Split(args, " ")...)
	//_, err := cmd.Output()
	res, err := execCmdWithOutput(netconfig.FlanneldPath, args)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
	return err
}

func runDockerOpt() error {
	args := fmt.Sprintf("-c")
	cmd := exec.Command(netconfig.MkDockerOptPath, strings.Split(args, " ")...)
	_, err := cmd.Output()
	return err
}

func ModifyDockerServiceConfig() error {
	file, err := os.OpenFile(netconfig.DockerServiceFilePath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("[boot] open file fail")
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	pos := int64(0)
	replace1 := "EnvironmentFile=-/run/docker_opts.env\n"
	replace2 := "ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock $DOCKER_OPTS\n"
	var flag = false
	for {
		//读取每一行文件内容
		line, Err := reader.ReadString('\n')
		if Err != nil {
			if Err == io.EOF {
				break
			} else {
				fmt.Println("read file error")
				return err
			}
		}
		if strings.Contains(line, "EnvironmentFile") {
			bytes := []byte(replace1)
			file.WriteAt(bytes, pos)
			flag = true
			pos += int64(len(replace1))
		} else if strings.Contains(line, "ExecStart") {
			if flag {
				//前一个改了，需要写一个
				bytes := []byte(replace2)
				file.WriteAt(bytes, pos)
			} else {
				//写两个
				bytes := []byte(replace1 + replace2)
				file.WriteAt(bytes, pos)
			}
			break
		} else {
			pos += int64(len(line))
		}
	}
	return nil
}
func restartDockerService() error {
	args := fmt.Sprintf("daemon-reload")
	cmd := exec.Command("systemctl", strings.Split(args, " ")...)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("[boot] restart Docker Service error")
		return err
	}
	args = fmt.Sprintf("restart docker")
	cmd = exec.Command("systemctl", strings.Split(args, " ")...)
	_, err = cmd.Output()
	return err
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
