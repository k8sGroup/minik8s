package autoscaler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/cmd/kubectl/app"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"time"
)

type scalableType string

const (
	replicaset scalableType = "replicaset"
	deployment scalableType = "deployment"
)

type scalableObject struct {
	kind scalableType
	key  string
}

type resourceStatus struct {
	metadata object.ObjectMeta
	memory   float64
	cpu      float64
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
	scaleInterval     time.Duration
	deploymentMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedDeployment]
	replicasetMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedReplicaset]
	autoscalerMap     *concurrentmap.ConcurrentMapTrait[string, object.VersionedAutoscaler]
	object2autoscaler *concurrentmap.ConcurrentMapTrait[scalableObject, stringAndChan] // mapping an object key to the autoscaler key
}

func NewAutoscalerController(ctx context.Context, controllerCtx util.ControllerContext) *AutoscalerController {
	promBase := fmt.Sprintf("http://%s:%s", controllerCtx.MasterIP, controllerCtx.PromServerPort)
	ac := &AutoscalerController{
		ls:                controllerCtx.Ls,
		promClient:        client.NewPromClient(promBase),
		stopChannel:       make(chan struct{}),
		resyncInterval:    time.Duration(controllerCtx.Config.ResyncIntervals) * time.Second,
		scaleInterval:     time.Duration(controllerCtx.Config.ScaleIntervals) * time.Second,
		deploymentMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedDeployment](),
		replicasetMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedReplicaset](),
		autoscalerMap:     concurrentmap.NewConcurrentMapTrait[string, object.VersionedAutoscaler](),
		object2autoscaler: concurrentmap.NewConcurrentMapTrait[scalableObject, stringAndChan](),
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
			err := acc.ls.Watch("/registry/rs/default", acc.watchReplicaset, acc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/rs : %s\n", err.Error())
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

	registerResyncLoop := func() {
		for {
			{
				resList, err := acc.ls.List("/registry/autoscaler/default")
				if err != nil {
					klog.Errorf("Error synchronizing!\n")
					goto failed
				}
				newMap := make(map[string]object.VersionedAutoscaler)
				for _, res := range resList {
					versionedAS := object.VersionedAutoscaler{Version: res.ResourceVersion}
					_ = json.Unmarshal(res.ValueBytes, &versionedAS.Autoscaler)
					newMap[res.Key] = versionedAS
				}
				acc.autoscalerMap.UpdateAll(newMap, object.SelectNewerAutoscaler)
			}
			{
				resList, err := acc.ls.List("/registry/rs/default")
				if err != nil {
					klog.Errorf("Error synchronizing!\n")
					goto failed
				}
				newMap := make(map[string]object.VersionedReplicaset)
				for _, res := range resList {
					versionedRS := object.VersionedReplicaset{Version: res.ResourceVersion}
					_ = json.Unmarshal(res.ValueBytes, &versionedRS.Replicaset)
					newMap[res.Key] = versionedRS
				}
				acc.replicasetMap.UpdateAll(newMap, object.SelectNewerReplicaset)
			}
			{
				resList, err := acc.ls.List("/registry/deployment/default")
				if err != nil {
					klog.Errorf("Error synchronizing!\n")
					goto failed
				}
				newMap := make(map[string]object.VersionedDeployment)
				for _, res := range resList {
					versionedDM := object.VersionedDeployment{Version: res.ResourceVersion}
					_ = json.Unmarshal(res.ValueBytes, &versionedDM.Deployment)
					newMap[res.Key] = versionedDM
				}
				acc.deploymentMap.UpdateAll(newMap, object.SelectNewerDeployment)
			}
		failed:
			time.Sleep(acc.resyncInterval)
		}
	}

	go registerResyncLoop()
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
		klog.Errorf("%s\n", err.Error())
		return
	}
	var ok bool
	var monitoringLoop func(stopCh <-chan struct{}, deploymentKey string)
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
	go monitoringLoop(stopChan, scalableObj.key)
}

func (acc *AutoscalerController) monitoringDeploymentLoop(stopCh <-chan struct{}, deploymentKey string) {
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
			pods, err := client.GetRSPods(acc.ls, vdm.Deployment.Metadata.Name)
		}
		time.Sleep(acc.scaleInterval)
	}
}

func (acc *AutoscalerController) monitoringReplicasetLoop(stopCh <-chan struct{}, replicasetKey string) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		vrs, ok := acc.replicasetMap.Get(replicasetKey)
		if ok {
			// TODO
			pods, err := client.GetRSPods(acc.ls, vrs.Replicaset.ObjectMeta.Name)
			if err != nil {
				goto ErrorHandle
			}
			statusList := acc.getPodsResourcePercentage(pods)
		}
	ErrorHandle:
		time.Sleep(acc.scaleInterval)
	}
}

func (acc *AutoscalerController) getPodsResourcePercentage(pods []*object.Pod) []resourceStatus {
	var statusList []resourceStatus
	for _, pod := range pods {
		var utilization *float64
		var err error
		var status resourceStatus
		if pod == nil {
			goto ErrorHandle
		}
		status.metadata = pod.ObjectMeta

		utilization, err = acc.promClient.GetResource(object.CPU_RESOURCE, pod.Spec.NodeName, pod.Name, nil)
		if err != nil || utilization == nil {
			goto ErrorHandle
		}
		status.cpu = *utilization

		utilization, err = acc.promClient.GetResource(object.MEMORY_RESOURCE, pod.Spec.NodeName, pod.Name, nil)
		if err != nil || utilization == nil {
			goto ErrorHandle
		}
		status.memory = *utilization

		statusList = append(statusList, status)

	ErrorHandle:
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
		key = "/registry/rs/default/" + targetObjName
		scalableObj.key = key
		scalableObj.kind = replicaset
		break
	default:
		return scalableObj, errors.New(fmt.Sprintf("wrong target obj kind [%s]", targetObjKind))
	}
	return scalableObj, nil
}
