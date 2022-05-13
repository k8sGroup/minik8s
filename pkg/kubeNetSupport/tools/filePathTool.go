package tools

import (
	"os"
	"strings"
)

const bootShPath = "/pkg/kubeNetSupport/boot/boot.sh"

//获取boot.sh文件的绝对路径
func GetBootShFilePath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	index := strings.Index(path, "minik8s")
	prefixString := path[:index+7]
	prefixString += bootShPath
	return prefixString, nil
}
