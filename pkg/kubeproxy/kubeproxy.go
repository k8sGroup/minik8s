package kubeproxy

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

type KubeProxy struct {
	ls              *listerwatcher.ListerWatcher
	Client          client.RESTClient
	dnsConfigWriter *DnsConfigWriter
	//etcd key到Svc Chain的映射, 一个service每个port对应一个svcChain
	ServiceName2SvcChain map[string]map[string]*SvcChain
	stopChannel          <-chan struct{}
}

func NewKubeProxy(lsConfig *listerwatcher.Config, clientConfig client.Config) *KubeProxy {
	res := &KubeProxy{}
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[kubeProxy] newKubeProxy Error")
		fmt.Println(err)
	}
	res.ls = ls
	res.Client = client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	res.stopChannel = make(chan struct{})
	res.ServiceName2SvcChain = make(map[string]map[string]*SvcChain)
	res.dnsConfigWriter = NewDnsConfigWriter(lsConfig, clientConfig)
	return res
}
func trans(from etcdstore.ListRes) etcdstore.WatchRes {
	return etcdstore.WatchRes{
		ResType:    etcdstore.PUT,
		Key:        from.Key,
		ValueBytes: from.ValueBytes,
	}
}
func (proxy *KubeProxy) PreSetService() {
	//拉取已经存在的service
	res, err := proxy.ls.List(config.ServicePrefix)
	if err != nil {
		fmt.Println("[kubeproxy]PreSetService error")
	} else {
		for _, val := range res {
			proxy.watchRuntimeService(trans(val))
		}
	}
}
func (proxy *KubeProxy) StartKubeProxy() {
	Boot()
	proxy.PreSetService()
	proxy.registry()
}

func (proxy *KubeProxy) registry() {
	//挂上watch， watch runtimeService
	watchService := func() {
		for {
			err := proxy.ls.Watch(config.ServicePrefix, proxy.watchRuntimeService, proxy.stopChannel)
			if err != nil {
				fmt.Println("[KubeProxy] watch error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go watchService()
}
func (proxy *KubeProxy) watchRuntimeService(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		svcs, ok := proxy.ServiceName2SvcChain[res.Key]
		if !ok {
			return
		} else {
			for _, v := range svcs {
				v.DeleteRule()
			}
			delete(proxy.ServiceName2SvcChain, res.Key)
		}
	} else {
		serviceRuntime := &object.Service{}
		err := json.Unmarshal(res.ValueBytes, serviceRuntime)
		fmt.Println(serviceRuntime)
		if err != nil {
			fmt.Println("[kubeProxy] Unmarshall fail")
			fmt.Println(err)
			return
		}
		svcS, ok := proxy.ServiceName2SvcChain[res.Key]
		if !ok {
			//先判断下service的state,决定是否创建service
			if serviceRuntime.Status.Phase == object.Failed {
				return
			}
			svcS = make(map[string]*SvcChain)
			for _, val := range serviceRuntime.Spec.Ports {
				var units []PodUnit
				for _, podNameAndIp := range serviceRuntime.Spec.PodNameAndIps {
					units = append(units, PodUnit{
						PodIp:   podNameAndIp.Ip,
						PodName: podNameAndIp.Name,
						PodPort: val.TargetPort,
					})
				}
				fmt.Println(units)
				tmp := NewSvcChain(serviceRuntime.MetaData.Name, NatTable, GeneralServiceChain, serviceRuntime.Spec.ClusterIp, val.Port, val.Protocol, units)
				tmp.ApplyRule()
				svcS[tmp.Name] = tmp
			}
			proxy.ServiceName2SvcChain[res.Key] = svcS
		} else {
			//更新
			for _, val := range serviceRuntime.Spec.Ports {
				var units []PodUnit
				for _, podNameAndIp := range serviceRuntime.Spec.PodNameAndIps {
					units = append(units, PodUnit{
						PodIp:   podNameAndIp.Ip,
						PodName: podNameAndIp.Name,
						PodPort: val.TargetPort,
					})
				}
				key := SvcChainPrefix + "-" + serviceRuntime.MetaData.Name + val.Port
				target, ok2 := svcS[key]
				if !ok2 {
					fmt.Println("[kubeProxy] Error, svc not found")
					return
				}
				fmt.Println(units)
				target.UpdateRule(units)
			}
		}
	}
}
