package service

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/netconfig"
	"sync"
	"time"
)

type Manager struct {
	//从service name 到 RuntimeService的映射
	serviceMap   map[string]*RuntimeService
	ls           *listerwatcher.ListerWatcher
	clientConfig client.Config
	stopChannel  chan struct{}
	client       client.RESTClient
	name2DnsMap  map[string]*object.DnsAndTrans
	lock         sync.Mutex
}

// Deprecated: Use NewServiceController and Manager.Run instead.
func NewManager(lsConfig *listerwatcher.Config, clientConfig client.Config) *Manager {
	manager := &Manager{}
	manager.serviceMap = make(map[string]*RuntimeService)
	manager.stopChannel = make(chan struct{})
	manager.name2DnsMap = make(map[string]*object.DnsAndTrans)
	var lock sync.Mutex
	manager.lock = lock
	manager.client = client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[Service Manager] newManager fail")
	}
	manager.ls = ls
	manager.clientConfig = clientConfig
	manager.register()
	go manager.checkAndBoot()
	go manager.checkDnsAndTrans()
	return manager
}

func NewServiceController(controllerCtx util.ControllerContext) *Manager {
	manager := &Manager{}
	manager.serviceMap = make(map[string]*RuntimeService)
	manager.stopChannel = make(chan struct{})
	manager.name2DnsMap = make(map[string]*object.DnsAndTrans)
	var lock sync.Mutex
	manager.lock = lock
	manager.client = client.RESTClient{
		Base: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}
	manager.ls = controllerCtx.Ls
	manager.clientConfig = client.Config{Host: controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort}
	return manager
}

func (manager *Manager) Run(ctx context.Context) {
	manager.register()
	go manager.checkAndBoot()
	go manager.checkDnsAndTrans()
	<-ctx.Done()
	close(manager.stopChannel)
}

//每隔一段时间check一下map中的DnsAndTrans, 看服务是否部署
func (manager *Manager) checkDnsAndTrans() {
	for {
		time.Sleep(2 * time.Second)
		manager.lock.Lock()
		var removes []string
		for k, v := range manager.name2DnsMap {
			resp, err := manager.client.GetRuntimeService(netconfig.GateWayServicePrefix + k)
			if err != nil {
				fmt.Println("[checkDnsAndTrans] getRuntimeService fail" + err.Error())
				continue
			}
			if resp == nil {
				continue
			}
			if resp.Status.Phase == object.Running {
				v.Status.Phase = object.ServiceCreated
				v.Spec.GateWayIp = resp.Spec.ClusterIp
				err = manager.client.UpdateDnsAndTrans(v)
				if err != nil {
					fmt.Println("[checkDnsAndService]updateDns fail" + err.Error())
					continue
				}
				removes = append(removes, k)
			}
		}
		for _, val := range removes {
			delete(manager.name2DnsMap, val)
		}
		manager.lock.Unlock()
	}
}

//没隔一段时间查看一下有无节点注册， 如果有注册的调用boot
func (manager *Manager) checkAndBoot() {
	for {
		time.Sleep(5 * time.Second)
		res, err := manager.ls.List(config.NODE_PREFIX)
		if err != nil {
			fmt.Println("[ServiceManager] checkAndBoot error" + err.Error())
			continue
		}
		if len(res) == 0 {
			continue
		} else {
			manager.boot()
			break
		}
	}
}
func (manager *Manager) boot() {
	//生成coreDns service
	err := manager.client.AddConfigRs(GetCoreDnsRsModule())
	if err != nil {
		fmt.Println("[ServiceManager] boot fail" + err.Error())
		return
	}
	time.Sleep(1 * time.Second)
	err = manager.client.UpdateService(GetCoreDnsServiceModule())
	if err != nil {
		fmt.Println("[ServiceManager] boot fail" + err.Error())
		return
	}
}
func (manager *Manager) register() {
	watchService := func() {
		for {
			err := manager.ls.Watch(config.ServiceConfigPrefix, manager.watchServiceConfig, manager.stopChannel)
			if err != nil {
				fmt.Println("[Service Manager] register error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}

	}
	watchDns := func() {
		for {
			err := manager.ls.Watch(config.DnsAndTransPrefix, manager.watchDnsAndTrans, manager.stopChannel)
			if err != nil {
				fmt.Println("[Server Manager] register error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
	go watchDns()
}
func (manager *Manager) watchDnsAndTrans(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		return
	}
	DnsAndTrans := &object.DnsAndTrans{}
	fmt.Println("[ServiceManager]watch Dns")
	err := json.Unmarshal(res.ValueBytes, DnsAndTrans)
	if err != nil {
		fmt.Println("[ServiceManager] Unmarshall error")
		return
	}
	if DnsAndTrans.Status.Phase == object.Delete {
		err = manager.client.DeleteService(netconfig.GateWayServicePrefix + DnsAndTrans.MetaData.Name)
		if err != nil {
			fmt.Println("[ServiceManager] watchDns: deleteService fail")
			fmt.Println(err)
			return
		}
		err = manager.client.DeleteConfigRs(netconfig.GateWayServicePrefix + DnsAndTrans.MetaData.Name)
		if err != nil {
			fmt.Println("[ServiceManager] watchDns: deleteRs fail")
			fmt.Println(err)
			return
		}
	} else if DnsAndTrans.Status.Phase == object.FileCreated {
		//需要生成rs以及service
		err = manager.client.AddConfigRs(GetGateWayRsModule(DnsAndTrans.MetaData.Name))
		if err != nil {
			fmt.Println("[ServiceManager] watchDns: addRs fail")
			fmt.Println(err)
			return
		}
		err = manager.client.UpdateService(GetGateWayServiceModule(DnsAndTrans.MetaData.Name))
		if err != nil {
			fmt.Println("[ServiceManager] watchDns: updateService fail")
			fmt.Println(err)
			return
		}
		//加入等待service部署的map
		manager.lock.Lock()
		manager.name2DnsMap[DnsAndTrans.MetaData.Name] = DnsAndTrans
		manager.lock.Unlock()
		//wait := 0
		//for {
		//	time.Sleep(1 * time.Second)
		//	resp, err2 := manager.client.GetRuntimeService(netconfig.GateWayServicePrefix + DnsAndTrans.MetaData.Name)
		//	if err2 != nil {
		//		fmt.Println("[serviceManager] wait for service deploy error" + err2.Error())
		//		break
		//	}
		//	if resp == nil {
		//		continue
		//	}
		//	if resp.Status.Phase == object.Running {
		//		DnsAndTrans.Status.Phase = object.ServiceCreated
		//		DnsAndTrans.Spec.GateWayIp = resp.Spec.ClusterIp
		//		err = manager.client.UpdateDnsAndTrans(DnsAndTrans)
		//		if err != nil {
		//			fmt.Println("[ServiceManager] updateDns fail" + err.Error())
		//		}
		//		break
		//	} else {
		//		fmt.Printf("wait round: %d", wait)
		//		wait++
		//		if wait > 60 {
		//			fmt.Println("[serviceManager] fail to deploy gateway service")
		//			break
		//		}
		//		continue
		//	}
		//}
	} else {
		//不用管这个
		//gateWay的更新通常是配置的更新，所以这里不用管, 用户端会更新配置文件
		return
	}

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
			manager.serviceMap[service.MetaData.Name] = NewRuntimeService(service, manager.ls, manager.clientConfig)
		} else {
			//修改service, 直接删了重新建一个
			runtimeService.DeleteService()
			delete(manager.serviceMap, service.MetaData.Name)
			manager.serviceMap[service.MetaData.Name] = NewRuntimeService(service, manager.ls, manager.clientConfig)
		}
	}
}
