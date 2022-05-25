package service

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/listerwatcher"
	"time"
)

type Manager struct {
	//从service name 到 RuntimeService的映射
	serviceMap   map[string]*RuntimeService
	ls           *listerwatcher.ListerWatcher
	lsConfig     *listerwatcher.Config
	clientConfig client.Config
	stopChannel  <-chan struct{}
	client       client.RESTClient
}

func NewManager(lsConfig *listerwatcher.Config, clientConfig client.Config) *Manager {
	manager := &Manager{}
	manager.serviceMap = make(map[string]*RuntimeService)
	manager.stopChannel = make(chan struct{})
	manager.client = client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[Service Manager] newManager fail")
	}
	manager.ls = ls
	manager.lsConfig = lsConfig
	manager.clientConfig = clientConfig
	manager.register()
	return manager
}

func (manager *Manager) register() {
	watchService := func() {
		for {
			err := manager.ls.Watch(config.ServiceConfigPrefix, manager.watchServiceConfig, manager.stopChannel)
			if err != nil {
				fmt.Println("[Service Manager] error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}

	}
	go watchService()
}
func (manager *Manager) watchServiceConfig(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		//不会有真删除的情况, 配置文件的删除通过设置status为DELETE
		return
	}
	service := &object.Service{}
	fmt.Println("[service manager]Watch receive")
	err := json.Unmarshal(res.ValueBytes, service)
	if err != nil {
		fmt.Println("[ServiceManager] Unmarshall error")
	}
	if service.Status.Phase == object.Delete {
		//需要删除service
		runtimeService, ok := manager.serviceMap[service.MetaData.Name]
		if !ok {
			return
		}
		runtimeService.DeleteService()
		delete(manager.serviceMap, service.MetaData.Name)
	} else {
		//可能是更新或者启动service
		runtimeService, ok := manager.serviceMap[service.MetaData.Name]
		if !ok {
			//新建service
			manager.serviceMap[service.MetaData.Name] = NewRuntimeService(service, manager.lsConfig, manager.clientConfig)
		} else {
			//修改service, 直接删了重新建一个
			runtimeService.DeleteService()
			delete(manager.serviceMap, service.MetaData.Name)
			manager.serviceMap[service.MetaData.Name] = NewRuntimeService(service, manager.lsConfig, manager.clientConfig)
		}
	}

}
