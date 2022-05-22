package service

import (
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/listerwatcher"
)

type RuntimeService struct {
	//service的配置文件
	serviceConfig *object.Service
	//service选择的Pod
	pods        []*object.Pod
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	Client      client.RESTClient
	Err         error
}


