package mesh

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	"strings"
	"sync"
	"time"
)

type EndPoint struct {
	PodIP  string
	Weight int
}

type Router struct {
	m      map[string][]EndPoint
	svcMap map[string]string // service name -> clusterIP
	mtx    sync.RWMutex

	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
}

func NewRouter() *Router {
	rand.Seed(time.Now().Unix())
	return &Router{}
}

func (d *Router) Run() {
	klog.Debugf("[ReplicaSetController]start running\n")
	go d.register()
	select {}
}

func (d *Router) register() {
	watchSvc := func(d *Router) {
		err := d.ls.Watch(config.ServicePrefix, d.watchRuntimeService, d.stopChannel)
		if err != nil {
			fmt.Printf("[Router] ListWatch init fail...")
		}
	}
	watchVirtualSvc := func(d *Router) {
		err := d.ls.Watch(config.VirtualSvcPrefix, d.watchVirtualService, d.stopChannel)
		if err != nil {
			fmt.Printf("[Router] ListWatch init fail...")
		}
	}
	go watchSvc(d)
	go watchVirtualSvc(d)
}

func (d *Router) watchRuntimeService(res etcdstore.WatchRes) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if res.ResType == etcdstore.DELETE {
		svcName := strings.TrimPrefix(res.Key, config.ServicePrefix+"/")
		fmt.Printf("[watchRuntimeService] delete svc:%v\n", svcName)
		clusterIP, ok := d.svcMap[svcName]
		if !ok {
			fmt.Printf("[watchRuntimeService] clusterIP cache not exist:%v\n", clusterIP)
			return
		}
		delete(d.m, clusterIP)
		delete(d.svcMap, svcName)
		return
	}

	svc := &object.Service{}
	err := json.Unmarshal(res.ValueBytes, svc)
	if err != nil {
		fmt.Println("[watchRuntimeService] Unmarshall fail")
		return
	}

	svcName := svc.MetaData.Name
	clusterIP := svc.Spec.ClusterIp
	d.svcMap[svcName] = clusterIP

	endpoints := d.m[clusterIP]
	weightMap := make(map[string]int)
	for _, ep := range endpoints {
		weightMap[ep.PodIP] = ep.Weight
	}

	newEndpoints := make([]EndPoint, 0)
	pods := svc.Spec.PodNameAndIps
	for _, pod := range pods {
		var weight int
		weight, ok := weightMap[pod.Ip]
		if !ok {
			weight = 0
		}
		newEndpoints = append(newEndpoints, EndPoint{pod.Ip, weight})
	}

	d.m[clusterIP] = newEndpoints
}

func (d *Router) watchVirtualService(res etcdstore.WatchRes) {
	vs := &object.VirtualService{}
	err := json.Unmarshal(res.ValueBytes, vs)
	if err != nil {
		fmt.Println("[watchVirtualService] Unmarshall fail")
		return
	}
	svcName := vs.Spec.Host
	clusterIP, ok := d.svcMap[svcName]
	if !ok {
		fmt.Printf("[watchVirtualService] service not exist:%v\n", svcName)
		return
	}

	vdest := vs.Spec.Route.VDest
	pdest := vs.Spec.Route.PDest

	if (len(vdest) == 0 && len(pdest) == 0) || (len(vdest) != 0 && len(pdest) != 0) {
		fmt.Printf("[watchVirtualService] invalid virtual service\n")
		return
	}

	//for _, route := range routes {
	//	route.
	//}
}

func (d *Router) UpsertEndpoints(clusterIP string, podIP string, weight int) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	endpoints, ok := d.m[clusterIP]

	if !ok {
		d.m[clusterIP] = []EndPoint{{podIP, weight}}
	} else {
		for _, ep := range endpoints {
			if ep.PodIP == podIP {
				ep.Weight = weight
				return
			}
		}
		d.m[clusterIP] = append(d.m[clusterIP], EndPoint{podIP, weight})
	}
}

func (d *Router) GetEndPoint(clusterIP string) (podIP *string, err error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	endpoints, ok := d.m[clusterIP]
	if !ok || len(endpoints) == 0 {
		return nil, errors.New("no endpoints")
	}

	var sum int
	for _, ep := range endpoints {
		sum += ep.Weight
	}

	num := rand.Intn(sum) + 1
	sum = 0
	for _, ep := range endpoints {
		sum += ep.Weight
		if sum >= num {
			return &ep.PodIP, nil
		}
	}

	return nil, errors.New("no endpoints chosen")
}
