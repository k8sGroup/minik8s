package main

import (
	"fmt"
	"minik8s/pkg/kubeNetSupport/boot"
	"os"
	"path"
	"path/filepath"
)

func GetProjectAbsPath() (projectAbsPath string) {
	programPath, _ := filepath.Abs(os.Args[0])
	fmt.Println("programPath:", programPath)
	projectAbsPath = path.Dir(path.Dir(programPath))
	fmt.Println("PROJECT_ABS_PATH:", projectAbsPath)
	return projectAbsPath
}
func main() {
	//err := boot.PreDownload()
	//path, err1 := tools.GetBootShFilePath()
	//if err1 != nil {
	//	fmt.Println(err1)
	//}
	err := boot.BootBasic("172.17.0.0/16")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("OK")
	}
}
