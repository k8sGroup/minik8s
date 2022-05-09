package util

import "minik8s/pkg/listerwatcher"

type ControllerContext struct {
	Ls             *listerwatcher.ListerWatcher
	MasterIP       string
	HttpServerPort string
}
