package main

import (
	"fmt"
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
	"os"
)

func main() {
	serverConfig := config.DefaultServerConfig()
	if len(os.Args) > 1 && os.Args[1] == "--recover" {
		fmt.Println("recover!")
		serverConfig.Recover = true
	}
	server, err := app.NewServer(serverConfig)
	if err != nil {
		panic(err)
	}
	err = server.Run()
	if err != nil {
		panic(err)
	}
}
