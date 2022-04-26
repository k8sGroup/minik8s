package main

import (
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/apiserver/config"
)

func main() {
	server, err := app.NewServer(config.DefaultServerConfig())
	if err != nil {
		panic(err)
	}
	err = server.Run()
	if err != nil {
		panic(err)
	}
}
