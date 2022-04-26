package main

import (
	"minik8s/cmd/kube-controller-manager/app"
	"minik8s/pkg/klog"
)

func main() {
	klog.Debugf("running controller manager.go\n")
	cmd := app.NewControllerManagerCommand()
	cmd.Execute()
}
