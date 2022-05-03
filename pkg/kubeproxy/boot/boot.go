package boot

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"minik8s/pkg/kubelet/dockerClient"
	"minik8s/pkg/kubeproxy/bridgeManager"
	"minik8s/pkg/kubeproxy/ipManager"
	"minik8s/pkg/kubeproxy/netconfig"
	"os"
	"os/exec"
	"strings"
)

//启动时配置主机网络
const DOCKER_CONFIG_PATH = "/etc/docker/daemon.json"

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

//修改配置文件
func modifyDocker0IpAndMask(ipAndMask string) error {
	_, err := os.Stat(DOCKER_CONFIG_PATH)
	if err != nil {
		//配置文件不存在
		f, err2 := os.Create(DOCKER_CONFIG_PATH)
		if err2 != nil {
			return err2
		}
		content := fmt.Sprintf("{\n \"bip\":\"%s\"\n}", ipAndMask)
		_, err2 = f.Write([]byte(content))
		if err2 != nil {
			return err2
		}
		f.Close()
	} else {
		//文件存在，更改bip项
		file, err2 := os.OpenFile(DOCKER_CONFIG_PATH, os.O_RDWR, 0666)
		if err2 != nil {
			return err2
		}
		reader := bufio.NewReader(file)
		pos := int64(0)
		flag := "\"bip\""
		for {
			line, err3 := reader.ReadString('\n')
			if err3 != nil {
				if err3 == io.EOF {
					//说明此时不存在bip字段
					break
				} else {
					file.Close()
					return err3
				}
			}
			if strings.Contains(line, flag) {
				content := fmt.Sprintf(" \"bip\":\"%s\"    ", ipAndMask)
				bytes := []byte(content)
				file.WriteAt(bytes, pos)
				file.Close()
				return nil
			}
			pos += int64(len(line))
		}
		//这种情况文件存在但是没有bip字段,直接覆盖
		f, err2 := os.Create(DOCKER_CONFIG_PATH)
		if err2 != nil {
			return err2
		}
		content := fmt.Sprintf("{\n \"bip\":\"%s\"\n}", ipAndMask)
		_, err2 = f.Write([]byte(content))
		if err2 != nil {
			return err2
		}
		f.Close()
	}
	return nil
}

//参数为本机docker0网段地址以及大网段地址(针对所有node)
//保证大网段涵盖所有node的docker网段
//例：172.17.43.1/24   172.17.0.0/16
func BootNetWork(Docker0IpAndMask string, BasicIpAndMask string) error {
	err := changeDocker0IpAndMask(Docker0IpAndMask)
	if err != nil {
		return err
	}
	err = preDownload()
	if err != nil {
		return err
	}
	err = createBr0()
	if err != nil {
		return err
	}
	err = bootBasic(BasicIpAndMask)
	if err != nil {
		return err
	}
	return nil
	//注意这时候还没有gre端口加入
}
func changeDocker0IpAndMask(ipAndMask string) error {
	//先停止所有的容器
	err := stopAllContainers()
	if err != nil {
		return err
	}
	//修改配置文件
	err = modifyDocker0IpAndMask(ipAndMask)
	if err != nil {
		return err
	}
	//命令行重启容器服务
	dockerCmd := "restart docker"
	cmd := exec.Command("systemctl", strings.Split(dockerCmd, " ")...)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

//建立基于ovs + gre的大二层通信
//先下载必要的软件
func preDownload() error {
	cmd := exec.Command("./pkg/kubeproxy/boot.sh")
	_, err := cmd.Output()
	return err
}
func createBr0() error {
	//创建br0网桥
	res, err := execOvsVsctlCmdWithOutput("list-br")
	if err != nil {
		return err
	}

	for _, value := range res {
		if strings.Contains(value, netconfig.OVS_BRIDGE_NAME) {
			//已经存在br0网桥，直接返回
			return nil
		}
	}
	//创建br0网桥
	command := "add-br br0"
	err = execOvsVsctlCmd(command)
	if err != nil {
		return err
	}
	return nil
}

////创建的基本配置启动
func bootBasic(BasicIpAndMask string) error {
	//将br0网桥加入到docker0网桥
	command := fmt.Sprintf("addif %s %s", netconfig.DOCKER_NETCARD, netconfig.OVS_BRIDGE_NAME)
	err := bridgeManager.ExecBrctlCmd(command)
	if err != nil {
		return err
	}
	err = ipManager.SetDev(netconfig.OVS_BRIDGE_NAME)
	if err != nil {
		return err
	}
	err = ipManager.SetDev(netconfig.DOCKER_NETCARD)
	if err != nil {
		return err
	}
	err = ipManager.AddRoute(BasicIpAndMask, netconfig.DOCKER_NETCARD)
	if err != nil {
		return err
	}
	return nil
}

//设置gre端口
func SetGrePortInBr0(grePort string, remoteIp string) error {
	//先判断grePort是否已经存在
	command := "list-ports " + netconfig.OVS_BRIDGE_NAME
	res, err := execOvsVsctlCmdWithOutput(command)
	if err != nil {
		return err
	}
	isExist := false
	for _, value := range res {
		if strings.Contains(value, grePort) {
			isExist = true
			break
		}
	}
	if isExist {
		//存在的情况下,先删除已存在的port
		command = "del-port " + grePort
		err = execOvsVsctlCmd(command)
		if err != nil {
			return err
		}
	}
	//创建gre Port
	command = fmt.Sprintf("add-port %s %s -- set Interface %s type=gre option:remote_ip=%s",
		netconfig.OVS_BRIDGE_NAME, grePort, grePort, remoteIp)
	err = execOvsVsctlCmd(command)
	if err != nil {
		return err
	}
	return nil
}
func delGrePortInBr0(grePort string) error {
	command := "del-port " + grePort
	err := execOvsVsctlCmd(command)
	return err
}

//ovs-vsctl cmd
func execOvsVsctlCmdWithOutput(command string) ([]string, error) {
	cmd := exec.Command("ovs-vsctl", strings.Split(command, " ")...)
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
func execOvsVsctlCmd(command string) error {
	cmd := exec.Command("ovs-vsctl", strings.Split(command, " ")...)
	_, err := cmd.Output()
	return err
}

//
