package main

import (
	"minik8s/cmd/kube-controller-manager/app"
)

func main() {
	command := app.NewControllerManagerCommand()
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}
