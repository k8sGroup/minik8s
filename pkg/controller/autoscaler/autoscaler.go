package autoscaler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/cmd/kubectl/app"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"path"
	"sync"
	"time"
)

type scalableType string

const (
	replicaset scalableType = "replicaset"
	deployment scalableType = "deployment"
)

type cpuPercentage float64
type memoryPercentage float64

type metricHandler struct {
	calculateFunc calculateStatus
	bound         float64
}

type calculateStatus func(statusList []resourceStatus) (cpuPercentage, memoryPercentage)

type scalableObject struct {
	kind scalableType
	key  string
}

type resourceStatus struct {
	metadata object.ObjectMeta
	memory   float64
	cpu      float64
}

func (s *resourceStatus) toString() string {
	return fmt.Sprintf("[Pod-Name] %-30s [Memory] %8.3f [CPU] %8.3f", s.metadata.Name, s.memory, s.cpu)
}

type stringAndChan struct {
	key    string
	stopCh chan<- struct{}
}

type AutoscalerController struct {
	ls                *listerwatcher.ListerWatcher
	promClient        *client.PromClient
	stopChannel       chan struct{}
	resyncInterval    time.Duration
	deploymentMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedDeployment]
	replicasetMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedReplicaset]
	autoscalerMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedAutoscaler]
	lockMap           *concurrentmap.ConcurrentMapTrait[string, sync.Mutex]
	object2autoscaler *concurrentmap.ConcurrentMapTrait[scalableObject, stringAndChan] // mapping an object key to the autoscaler key
	apiServerBase     string
}

func NewAutoscalerController(ctx context.Context, controllerCtx util.ControllerContext) *AutoscalerController {
	promBase := fmt.Sprintf("http://%s:%s", controllerCtx.MasterIP, controllerCtx.PromServerPort)
	ac := &AutoscalerController{
		ls:                controllerCtx.Ls,
		promClient:        client.NewPromClient(promBase),
		stopChannel:       make(chan struct{}),
		resyncInterval:    time.Duration(controllerCtx.Config.ResyncIntervals) * time.Second,
		deploymentMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedDeployment](),
		replicasetMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedReplicaset](),
		autoscalerMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedAutoscaler](),
		lockMap:           concurrentmap.NewConcurrentMapTrait[string, sync.Mutex](),
		object2autoscaler: concurrentmap.NewConcurrentMapTrait[scalableObject, stringAndChan](),
		apiServerBase:     "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}
	return ac
}

func (acc *AutoscalerController) Run(ctx context.Context) {
	acc.register()
	<-ctx.Done()
	close(acc.stopChannel)
}

func (acc *AutoscalerController) register() {
	registerWatchAutoscaler := func() {
		for {
			err := acc.ls.Watch("/registry/autoscaler/default", acc.watchAutoscaler, acc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/autoscaler : %s\n", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerWatchReplicaset := func() {
		for {
			err := acc.ls.Watch(config.RSConfigPrefix, acc.watchReplicaset, acc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/rsConfig : %s\n", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerWatchDeployment := func() {
		for {
			err := acc.ls.Watch("/registry/deployment/default", acc.watchDeployment, acc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment : %s\n", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}

	//registerResyncLoop := func() {
	//	for {
	//		{
	//			resList, err := acc.ls.List("/registry/autoscaler/default")
	//			if err != nil {
	//				klog.Errorf("Error synchronizing!\n")
	//				goto failed
	//			}
	//			newMap := make(map[string]object.VersionedAutoscaler)
	//			for _, res := range resList {
	//				versionedAS := object.VersionedAutoscaler{Version: res.ResourceVersion}
	//				_ = json.Unmarshal(res.ValueBytes, &versionedAS.Autoscaler)
	//				newMap[res.Key] = versionedAS
	//			}
	//			acc.autoscalerMap.UpdateAll(newMap, object.SelectNewerAutoscaler)
	//		}
	//		{
	//			resList, err := acc.ls.List(config.RSConfigPrefix)
	//			if err != nil {
	//				klog.Errorf("Error synchronizing!\n")
	//				goto failed
	//			}
	//			newMap := make(map[string]object.VersionedReplicaset)
	//			for _, res := range resList {
	//				versionedRS := object.VersionedReplicaset{Version: res.ResourceVersion}
	//				_ = json.Unmarshal(res.ValueBytes, &versionedRS.Replicaset)
	//				newMap[res.Key] = versionedRS
	//			}
	//			acc.replicasetMap.UpdateAll(newMap, object.SelectNewerReplicaset)
	//		}
	//		{
	//			resList, err := acc.ls.List("/registry/deployment/default")
	//			if err != nil {
	//				klog.Errorf("Error synchronizing!\n")
	//				goto failed
	//			}
	//			newMap := make(map[string]object.VersionedDeployment)
	//			for _, res := range resList {
	//				versionedDM := object.VersionedDeployment{Version: res.ResourceVersion}
	//				_ = json.Unmarshal(res.ValueBytes, &versionedDM.Deployment)
	//				newMap[res.Key] = versionedDM
	//			}
	//			acc.deploymentMap.UpdateAll(newMap, object.SelectNewerDeployment)
	//		}
	//	failed:
	//		time.Sleep(acc.resyncInterval)
	//	}
	//}
	//go registerResyncLoop()

	go registerWatchAutoscaler()
	go registerWatchDeployment()
	go registerWatchReplicaset()
}

/*
Autoscaler will take control of object.Deployment or object.ReplicaSet after it is applied.

Deleting an object.Autoscaler will not delete its deployments or replica-sets.

Deleting an object.Deployment or object.ReplicaSet will make its owner autoscaler unavailable.
*/
func (acc *AutoscalerController) watchAutoscaler(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		ac := object.Autoscaler{}
		err := json.Unmarshal(res.ValueBytes, &ac)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		vac := object.VersionedAutoscaler{
			Version:    res.ResourceVersion,
			Autoscaler: ac,
		}
		acc.handleAutoscalerPut(res.Key, vac)
		break
	case etcdstore.DELETE:
		acc.handleAutoscalerDel(res.Key)
		break
	}
}

func (acc *AutoscalerController) watchDeployment(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		dm := object.Deployment{}
		err := json.Unmarshal(res.ValueBytes, &dm)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		vdm := object.VersionedDeployment{
			Version:    res.ResourceVersion,
			Deployment: dm,
		}
		acc.deploymentMap.Put(res.Key, vdm)
		break
	case etcdstore.DELETE:
		scalableObj := scalableObject{
			kind: deployment,
			key:  res.Key,
		}
		acc.handleScalableObjectDel(scalableObj)
		break
	}
}

func (acc *AutoscalerController) watchReplicaset(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		rs := object.ReplicaSet{}
		err := json.Unmarshal(res.ValueBytes, &rs)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		if rs.Spec.Replicas == 0 {
			acc.replicasetMap.Del(res.Key)
		} else {
			vrs := object.VersionedReplicaset{
				Version:    res.ResourceVersion,
				Replicaset: rs,
			}
			acc.replicasetMap.Put(res.Key, vrs)
		}
		break
	case etcdstore.DELETE:
		scalableObj := scalableObject{
			kind: replicaset,
			key:  res.Key,
		}
		acc.handleScalableObjectDel(scalableObj)
		break
	}
}

func (acc *AutoscalerController) handleAutoscalerPut(autoscalerKey string, vac object.VersionedAutoscaler) {
	scalableObj, err := autoscaler2ScalableObject(vac)
	if err != nil {
		return
	}
	calculateFuncMap := make(map[string]metricHandler)
	for _, metric := range vac.Autoscaler.Spec.Metrics {
		if metric.Name == object.MetricCPU || metric.Name == object.MetricMemory {
			switch metric.Strategy {
			case object.MetricMax:
				calculateFuncMap[metric.Name] = metricHandler{
					calculateFunc: calculateMaxStatus,
					bound:         float64(metric.Percentage) / 100,
				}
				break
			case object.MetricAverage:
				calculateFuncMap[metric.Name] = metricHandler{
					calculateFunc: calculateAverageStatus,
					bound:         float64(metric.Percentage) / 100,
				}
				break
			default:
				break
			}
		}
	}
	fmt.Println("Init monitoring loop")
	var ok bool
	var monitoringLoop func(stopCh <-chan struct{}, deploymentKey string, calculateFuncMap map[string]metricHandler, maxReplicas int32, minReplicas int32, scaleInterval int32)
	switch scalableObj.kind {
	case deployment:
		_, ok = acc.deploymentMap.Get(scalableObj.key)
		if !ok {
			return
		} else {
			monitoringLoop = acc.monitoringDeploymentLoop
		}
		break
	case replicaset:
		_, ok = acc.replicasetMap.Get(scalableObj.key)
		if !ok {
			fmt.Println("replicaset not found, return !")
			return
		} else {
			monitoringLoop = acc.monitoringReplicasetLoop
		}
		break
	default:
		return
	}
	stopChan := make(chan struct{})
	autoscalerMeta := stringAndChan{
		key:    autoscalerKey,
		stopCh: stopChan,
	}
	acc.autoscalerMap.Put(autoscalerKey, vac)
	acc.object2autoscaler.Put(scalableObj, autoscalerMeta)
	var interval int32 = 10
	if vac.Autoscaler.Spec.ScaleInterval > 0 {
		interval = vac.Autoscaler.Spec.ScaleInterval
	}
	go monitoringLoop(stopChan, scalableObj.key, calculateFuncMap, vac.Autoscaler.Spec.MaxReplicas, vac.Autoscaler.Spec.MinReplicas, interval)
}

func (acc *AutoscalerController) monitoringDeploymentLoop(stopCh <-chan struct{}, deploymentKey string, calculateFuncMap map[string]metricHandler, maxReplicas int32, minReplicas int32, interval int32) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		vdm, ok := acc.deploymentMap.Get(deploymentKey)
		if ok {
			/*
				The reason for using deployment here is that the replicaset which is controlled by a deployment
				has the same name as its owner deployment's name.
				And at a particular point in time, one deployment may have more than one replicaset.
				They all have the same name but different keys.
			*/
			rs, rsExist := getControlledRS(vdm.Deployment, acc.replicasetMap.SnapShot())
			if !rsExist {
				goto StepEnd
			}
			pods, err := client.GetRSPods(acc.ls, rs.ObjectMeta.Name, rs.UID)
			if err != nil {
				goto StepEnd
			}
			podStatusList := acc.getPodsResourcePercentage(pods)

			var cpu, cpuBound cpuPercentage
			var memory, memoryBound memoryPercentage
			var cpuMetric, memoryMetric bool
			var handler metricHandler

			if handler, cpuMetric = calculateFuncMap[object.MetricCPU]; cpuMetric {
				cpu, _ = handler.calculateFunc(podStatusList)
				cpuBound = cpuPercentage(handler.bound)
			}
			if handler, memoryMetric = calculateFuncMap[object.MetricMemory]; memoryMetric {
				_, memory = handler.calculateFunc(podStatusList)
				memoryBound = memoryPercentage(handler.bound)
			}

			rsKey := path.Join(config.RSConfigPrefix, rs.Name+rs.UID)
			if (cpuMetric && cpu > cpuBound) || (memoryMetric && memory > memoryBound) {
				rs.Spec.Replicas += 1
				if rs.Spec.Replicas <= maxReplicas {
					err = client.Put(acc.apiServerBase+rsKey, rs)
					if err != nil {
						goto StepEnd
					}
				}
			} else if (cpuMetric && memoryMetric && cpu < cpuBound && memory < memoryBound) || (cpuMetric && !memoryMetric && cpu < cpuBound) || (memoryMetric && !cpuMetric && memory < memoryBound) {
				rs.Spec.Replicas -= 1
				if rs.Spec.Replicas >= minReplicas {
					err = client.Put(acc.apiServerBase+rsKey, rs)
					if err != nil {
						goto StepEnd
					}
				}
			}
		}
	StepEnd:
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (acc *AutoscalerController) monitoringReplicasetLoop(stopCh <-chan struct{}, replicasetKey string, calculateFuncMap map[string]metricHandler, maxReplicas int32, minReplicas int32, interval int32) {
	fmt.Println("monitoring replicaset loop")
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		vrs, ok := acc.replicasetMap.Get(replicasetKey)
		if ok {
			pods, err := client.GetRSPods(acc.ls, vrs.Replicaset.ObjectMeta.Name, vrs.Replicaset.UID)
			if err != nil {
				fmt.Printf("cannot get pods of replicaset %s\n", vrs.Replicaset.ObjectMeta.Name)
				goto StepEnd
			}
			podStatusList := acc.getPodsResourcePercentage(pods)
			fmt.Println("pod's status information")
			for _, status := range podStatusList {
				fmt.Println(status.toString())
			}

			var cpu, cpuBound cpuPercentage
			var memory, memoryBound memoryPercentage
			var cpuMetric, memoryMetric bool
			var handler metricHandler

			if handler, cpuMetric = calculateFuncMap[object.MetricCPU]; cpuMetric {
				cpu, _ = handler.calculateFunc(podStatusList)
				cpuBound = cpuPercentage(handler.bound)
			}
			if handler, memoryMetric = calculateFuncMap[object.MetricMemory]; memoryMetric {
				_, memory = handler.calculateFunc(podStatusList)
				memoryBound = memoryPercentage(handler.bound)
			}

			mtx := acc.lockMap.PutIfNotExist(replicasetKey, sync.Mutex{})

			func() {
				if mtx.TryLock() {
					defer mtx.Unlock()
					fmt.Printf("[CPU BOUND] %5.2f [CPU] %5.2f [MEM BOUND] %5.2f [MEM] %5.2f\n", cpuBound, cpu, memoryBound, memory)
					fmt.Printf("check cpu ? %t\tcheck mem ? %t\n", cpuMetric, memoryMetric)
					if (cpuMetric && cpu > cpuBound) || (memoryMetric && memory > memoryBound) {
						fmt.Println("increase replicas")
						for vrs.Replicaset.Spec.Replicas < maxReplicas {
							vrs.Replicaset.Spec.Replicas += 1
							err = client.Put(acc.apiServerBase+replicasetKey, vrs.Replicaset)
							if err != nil {
								vrs.Replicaset.Spec.Replicas -= 1
							}
							time.Sleep(time.Duration(interval) * time.Second)
						}
					} else if (cpuMetric && memoryMetric && cpu < cpuBound && memory < memoryBound) || (cpuMetric && !memoryMetric && cpu < cpuBound) || (memoryMetric && !cpuMetric && memory < memoryBound) {
						fmt.Println("decrease replicas")
						for vrs.Replicaset.Spec.Replicas > minReplicas {
							vrs.Replicaset.Spec.Replicas -= 1
							err = client.Put(acc.apiServerBase+replicasetKey, vrs.Replicaset)
							if err != nil {
								vrs.Replicaset.Spec.Replicas += 1
							}
							time.Sleep(time.Duration(interval) * time.Second)
						}
					}
				}
			}()
		}
	StepEnd:
		time.Sleep(time.Second)
	}
}

func getControlledRS(deployment object.Deployment, rsMap map[string]object.VersionedReplicaset) (object.ReplicaSet, bool) {
	versionedRS := object.VersionedReplicaset{
		Version: 0,
	}
	for _, vrs := range rsMap {
		for _, owner := range vrs.Replicaset.OwnerReferences {
			if owner.UID == deployment.Metadata.UID && owner.Name == deployment.Metadata.Name && vrs.Version >= versionedRS.Version {
				versionedRS = vrs
			}
		}
	}
	if versionedRS.Version == 0 {
		return versionedRS.Replicaset, false
	} else {
		return versionedRS.Replicaset, true
	}
}

func calculateMaxStatus(statusList []resourceStatus) (cpuPercentage, memoryPercentage) {
	var cpu float64 = 0
	var memory float64 = 0
	for _, status := range statusList {
		cpu = math.Max(cpu, status.cpu)
		memory = math.Max(memory, status.memory)
	}
	return cpuPercentage(cpu), memoryPercentage(memory)
}

func calculateAverageStatus(statusList []resourceStatus) (cpuPercentage, memoryPercentage) {
	var cpu float64 = 0
	var memory float64 = 0
	length := len(statusList)
	if length == 0 {
		return 0, 0
	}
	for _, status := range statusList {
		cpu += status.cpu
		memory += status.memory
	}
	return cpuPercentage(cpu / float64(length)), memoryPercentage(memory / float64(length))
}

func (acc *AutoscalerController) getPodsResourcePercentage(pods []*object.Pod) []resourceStatus {
	var statusList []resourceStatus
	for _, pod := range pods {
		var utilization *float64
		var err error
		var status resourceStatus
		if pod == nil {
			goto StepEnd
		}
		status.metadata = pod.ObjectMeta

		utilization, err = acc.promClient.GetResource(object.CPU_RESOURCE, pod.Name, pod.UID, nil)
		if err != nil || utilization == nil {
			goto StepEnd
		}
		status.cpu = *utilization

		utilization, err = acc.promClient.GetResource(object.MEMORY_RESOURCE, pod.Name, pod.UID, nil)
		if err != nil || utilization == nil {
			goto StepEnd
		}
		status.memory = *utilization

		statusList = append(statusList, status)

	StepEnd:
		continue
	}
	return statusList
}

func (acc *AutoscalerController) handleAutoscalerDel(key string) {
	ac, ok := acc.autoscalerMap.Get(key)
	if !ok {
		return
	}
	acc.autoscalerMap.Del(key)
	scalableObj, err := autoscaler2ScalableObject(ac)
	if err != nil {
		klog.Errorf("%s\n", err.Error())
		return
	}
	autoscalerMeta, ok := acc.object2autoscaler.Get(scalableObj)
	if !ok {
		return
	}
	acc.object2autoscaler.Del(scalableObj)
	close(autoscalerMeta.stopCh)
}

func (acc *AutoscalerController) handleScalableObjectDel(scalableObj scalableObject) {
	switch scalableObj.kind {
	case replicaset:
		acc.replicasetMap.Del(scalableObj.key)
		break
	case deployment:
		acc.deploymentMap.Del(scalableObj.key)
		break
	default:
		return
	}
	autoscalerMeta, ok := acc.object2autoscaler.Get(scalableObj)
	if !ok {
		return
	}
	acc.object2autoscaler.Del(scalableObj)
	acc.autoscalerMap.Del(scalableObj.key)
	close(autoscalerMeta.stopCh)
}

func autoscaler2ScalableObject(vac object.VersionedAutoscaler) (scalableObject, error) {
	targetObjKind := vac.Autoscaler.Spec.ScaleTargetRef.Kind
	targetObjName := vac.Autoscaler.Spec.ScaleTargetRef.Name
	scalableObj := scalableObject{}
	key := ""
	switch targetObjKind {
	case app.Deployment:
		key = "/registry/deployment/default/" + targetObjName
		scalableObj.key = key
		scalableObj.kind = deployment
		break
	case app.Replicaset:
		key = path.Join(config.RSConfigPrefix, targetObjName)
		scalableObj.key = key
		scalableObj.kind = replicaset
		break
	default:
		return scalableObj, errors.New(fmt.Sprintf("wrong target obj kind [%s]", targetObjKind))
	}
	return scalableObj, nil
}
