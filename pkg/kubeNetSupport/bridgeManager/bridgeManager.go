package bridgeManager

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
)

func ExecBrctlCmd(command string) error {
	cmd := exec.Command("brctl", strings.Split(command, " ")...)
	_, err := cmd.Output()
	return err
}
func ExecBrctlCmdWithOutput(command string) ([]string, error) {
	cmd := exec.Command("brctl", strings.Split(command, " ")...)
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
