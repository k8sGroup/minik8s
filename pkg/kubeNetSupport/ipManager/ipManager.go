package ipManager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

//修改路由表的工具函数
func GetRouteInfo() ([]string, error) {
	command := "route show"
	return execIpCmdWithOutput(command)
}
func AddRoute(ipAndMask string, dev string) error {
	command := fmt.Sprintf("route add %s dev %s", ipAndMask, dev)
	return execIpCmd(command)
}
func DelRoute(ipAndMask string, dev string) error {
	command := fmt.Sprintf("route del %s dev %s", ipAndMask, dev)
	return execIpCmd(command)
}

//启动设备
func SetDev(dev string) error {
	command := fmt.Sprintf("link set dev %s up", dev)
	return execIpCmd(command)
}
func execIpCmd(iptableCmd string) error {
	cmd := exec.Command("ip", strings.Split(iptableCmd, " ")...)
	_, err := cmd.Output()
	return err
}
func execIpCmdWithOutput(iptableCmd string) ([]string, error) {
	cmd := exec.Command("ip", strings.Split(iptableCmd, " ")...)
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
