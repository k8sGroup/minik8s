package iptablesManager

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"minik8s/pkg/klog"
	"os"
	"os/exec"
	"strings"
)

type PortMapping struct {
	DockerIp   string
	DockerPort string
	HostPort   string
}

//修改路由表的工具函数
//Docker chain 的编辑
//建立端口映射, 使得外界可以以 hostIp + hostPort的形式访问到docker
func AddDockerChainMappingRule(dockerIp string, dockerPort string, hostPort string) (PortMapping, error) {
	iptableCmd := fmt.Sprintf("-t nat -A DOCKER -p tcp --dport %s -j DNAT --to-destination %s:%s",
		hostPort, dockerIp, dockerPort)
	cmd := exec.Command("iptables", strings.Split(iptableCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		klog.Errorf("iptables output err: %s -- %v", output, err)
		return PortMapping{}, err
	}
	return PortMapping{
		DockerPort: dockerPort,
		DockerIp:   dockerIp,
		HostPort:   hostPort,
	}, err
}
func RemoveDockerChainMappingRule(portMapping PortMapping) error {
	iptableCmd := fmt.Sprintf("-t nat -vnL %s --line-number", "DOCKER")
	values, err := execIptablesCmdWithOutput(iptableCmd)
	if err != nil {
		return err
	}
	//找具体在哪行
	flag1 := "dpt:" + portMapping.HostPort
	flag2 := "to:" + portMapping.DockerIp + ":" + portMapping.DockerPort
	for _, value := range values {
		if strings.Index(value, flag1) != -1 && strings.Index(value, flag2) != -1 {
			//获取rule行号
			num := getRuleSequenceNum(value)
			cmd := "-t nat -D DOCKER " + num
			err = execIptablesCmd(cmd)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}
func getRuleSequenceNum(input string) string {
	begin := 0
	for {
		if input[begin] >= '0' && input[begin] <= '9' {
			break
		}
		begin++
	}
	end := begin + 1
	for {
		if input[end] < '0' || input[end] > '9' {
			break
		}
		end++
	}
	return input[begin:end]
}
func getPortMappingFromString(input string) (PortMapping, error) {
	index := strings.Index(input, "dpt:")
	if index == -1 {
		return PortMapping{}, errors.New("failed")
	}
	index += 4
	end := index + 1
	for {
		if input[end] > '9' || input[end] < '0' {
			break
		}
		end++
	}
	hostPort := input[index:end]
	index = strings.Index(input, "to:")
	if index == -1 {
		return PortMapping{}, errors.New("failed")
	}
	index += 3
	ipAndPort := input[index:]
	index = strings.Index(ipAndPort, ":")
	if index == -1 {
		return PortMapping{}, errors.New("failed")
	}
	ip := ipAndPort[0:index]
	port := ipAndPort[index+1:]
	return PortMapping{
		HostPort:   hostPort,
		DockerIp:   ip,
		DockerPort: port,
	}, nil
}

func GetNatDockerChainRules() ([]PortMapping, error) {
	iptableCmd := fmt.Sprintf("-t nat -vnL %s --line-number", "DOCKER")
	values, err := execIptablesCmdWithOutput(iptableCmd)
	if err != nil {
		return nil, err
	}
	var result []PortMapping
	for _, value := range values {
		tmp, err2 := getPortMappingFromString(value)
		if err2 == nil {
			result = append(result, tmp)
		}
	}
	return result, nil
}
func execIptablesCmd(iptableCmd string) error {
	cmd := exec.Command("iptables", strings.Split(iptableCmd, " ")...)
	_, err := cmd.Output()
	return err
}
func execIptablesCmdWithOutput(iptableCmd string) ([]string, error) {
	cmd := exec.Command("iptables", strings.Split(iptableCmd, " ")...)
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
