package main

import (
	"fmt"
	"minik8s/pkg/client"
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
	clientConfig := client.Config{Host: "192.168.1.7:8080"}
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	ress, err := client.Get(restClient.Base + "/registry/node/default")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, res := range ress {
			fmt.Println(string(res.ValueBytes), res.Key)
		}

	}
}
