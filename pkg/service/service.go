package service

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/listerwatcher"
	"sync"
)

type RuntimeService struct {
	//service的配置文件
	serviceConfig *object.Service
	//service选择的Pod
	pods        []*object.Pod
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	Client      client.RESTClient
	rwLock      sync.RWMutex
	Err         error
}

type ManagerOfService struct {
	//从service name 到 RuntimeService的映射
	serviceMap  map[string]*RuntimeService
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	client      client.RESTClient
}

func (manager *ManagerOfService) register() {

	watchServiceConfig := func() {
		for {
			err := manager.ls.Watch(config.ServiceConfigPrefix, manager.watchServiceConfig, manager.stopChannel)
		}
	}
}
func (manager *ManagerOfService) watchServiceConfig(res etcdstore.WatchRes) {
	//进行预处理然后创建一个RuntimeService
	switch res.ResType {
	case etcdstore.PUT:
		service := &object.Service{}
		err := json.Unmarshal(res.ValueBytes, service)
		if err != nil {
			fmt.Println("[ManagerOfService] Unmarshall error")
		}
	}
}
