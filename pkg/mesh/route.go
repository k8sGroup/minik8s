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

func NewRouter(lsConfig *listerwatcher.Config) *Router {
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[NewRouter] list watch fail...")
	}
	m := make(map[string][]EndPoint)
	svcMap := make(map[string]string)
	return &Router{
		ls:     ls,
		m:      m,
		svcMap: svcMap,
	}
}

func (d *Router) Run() {
	rand.Seed(time.Now().Unix())
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

	fmt.Printf("[watchRuntimeService] service:%+v\n", svc)

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

	fmt.Printf("[watchRuntimeService] clusterIP:%+v endpoints:%+v\n", clusterIP, d.m[clusterIP])
}

func (d *Router) watchVirtualService(res etcdstore.WatchRes) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	vs := &object.VirtualService{}
	err := json.Unmarshal(res.ValueBytes, vs)
	if err != nil {
		fmt.Println("[watchVirtualService] Unmarshall fail")
		return
	}
	clusterIP := vs.Spec.Host

	pdest := vs.Spec.Route.PDest

	if len(pdest) != 0 {
		for _, pod := range pdest {
			podIP := pod.PodIP
			weight := pod.Weight
			d.UpsertEndpoints(clusterIP, podIP, int(weight))
		}
	}

	endpoints, _ := d.m[clusterIP]
	fmt.Printf("[watchVirtualService] Update weight:%v\n", endpoints)
}

func (d *Router) UpsertEndpoints(clusterIP string, podIP string, weight int) {

	endpoints, ok := d.m[clusterIP]
	if !ok {
		return
	}

	newEndPoints := []EndPoint{{podIP, weight}}

	for _, ep := range endpoints {
		if ep.PodIP == podIP {
			continue
		}
		newEndPoints = append(newEndPoints, ep)
	}

	d.m[clusterIP] = newEndPoints
}

func (d *Router) GetEndPoint(clusterIP string, direction string) (podIP *string, err error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	endpoints, ok := d.m[clusterIP]
	if !ok {
		return &clusterIP, nil
	} else if len(endpoints) == 0 {
		fmt.Printf("[Endpoint:%v] endpoints for service not exist:%v\n", direction, clusterIP)
		return nil, errors.New("no endpoints")
	}

	var sum int
	for _, ep := range endpoints {
		sum += ep.Weight
	}

	if sum == 0 {
		idx := rand.Intn(len(endpoints))
		fmt.Printf("[Endpoint:%v] %v for service %v\n", direction, endpoints[idx].PodIP, clusterIP)
		return &endpoints[idx].PodIP, nil
	}

	num := rand.Intn(sum) + 1
	sum = 0
	for _, ep := range endpoints {
		sum += ep.Weight
		if sum >= num {
			fmt.Printf("[Endpoint:%v] %v for service %v\n", direction, ep.PodIP, clusterIP)
			return &ep.PodIP, nil
		}
	}

	fmt.Printf("[GetEndPoint] find enpoint by weight fail, clusterIP:%v\n", clusterIP)

	return nil, errors.New("no endpoints chosen")
}
