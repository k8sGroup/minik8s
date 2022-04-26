package main

import (
	"minik8s/cmd/kube-controller-manager/app"
	"minik8s/pkg/klog"
)

func main() {
	klog.Infof("running controller manager.go\n")
	command := app.NewControllerManagerCommand()
	err := command.Execute()
	if err != nil {
		klog.Errorf("controller command fail")
	}
}
