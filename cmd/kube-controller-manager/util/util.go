package util

import (
	"minik8s/cmd/kube-controller-manager/app/config"
	"minik8s/pkg/listerwatcher"
)

type ControllerContext struct {
	Ls             *listerwatcher.ListerWatcher
	MasterIP       string
	HttpServerPort string
	PromServerPort string
	Config         *config.CompletedConfig
}
