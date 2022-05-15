package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"time"
)

type DeploymentController struct {
	ls             *listerwatcher.ListerWatcher
	deploymentMap  *concurrentmap.ConcurrentMapTrait[string, object.VersionedDeployment]
	replicasetMap  *concurrentmap.ConcurrentMapTrait[string, object.VersionedReplicaset]
	dm2rs          *concurrentmap.ConcurrentMapTrait[string, string] // dm2rs mapping from Deployment key in etcdstore to Replicaset key in etcdstore
	resyncInterval time.Duration
	stopChannel    chan struct{}
	apiServerBase  string
}

func NewDeploymentController(ctx context.Context, controllerCtx util.ControllerContext) *DeploymentController {
	dc := &DeploymentController{
		ls:             controllerCtx.Ls,
		deploymentMap:  concurrentmap.NewConcurrentMapTrait[string, object.VersionedDeployment](),
		replicasetMap:  concurrentmap.NewConcurrentMapTrait[string, object.VersionedReplicaset](),
		dm2rs:          concurrentmap.NewConcurrentMapTrait[string, string](),
		stopChannel:    make(chan struct{}),
		resyncInterval: time.Duration(controllerCtx.Config.ResyncIntervals) * time.Second,
		apiServerBase:  "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}
	if dc.apiServerBase == "" {
		klog.Fatalf("uninitialized apiserver base!\n")
	}
	return dc
}

func (dc *DeploymentController) Run(ctx context.Context) {
	klog.Debugf("[DeploymentController] running...\n")
	dc.register()
	<-ctx.Done()
	close(dc.stopChannel)
}

func (dc *DeploymentController) register() {
	registerWatchReplicaset := func() {
		for {
			err := dc.ls.Watch("/registry/rs/default", dc.watchReplicaset, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/rs\n")
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerAddDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment/default", dc.putDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment : %s\n", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerDeleteDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment/default", dc.deleteDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment : %s\n", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerResyncLoop := func() {
		for {
			{
				resList, err := dc.ls.List("/registry/rs/default")
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
				dc.replicasetMap.UpdateAll(newMap, object.SelectNewerReplicaset)
			}
			{
				resList, err := dc.ls.List("/registry/deployment/default")
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
				dc.deploymentMap.UpdateAll(newMap, object.SelectNewerDeployment)
			}
		failed:
			time.Sleep(dc.resyncInterval)
		}
	}

	go registerResyncLoop()
	go registerAddDeployment()
	go registerDeleteDeployment()
	go registerWatchReplicaset()
}

func (dc *DeploymentController) putDeployment(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.PUT {
		return
	}
	deployment := object.Deployment{}
	err := json.Unmarshal(res.ValueBytes, &deployment)
	if err != nil {
		klog.Errorf("%s\n", err.Error())
		return
	}
	if res.IsCreate {
		rsKeyNew := "/registry/rs/default/" + deployment.Metadata.Name + uuid.New().String()
		rs := object.ReplicaSet{
			ObjectMeta: object.ObjectMeta{
				Name:   deployment.Metadata.Name,
				Labels: deployment.Metadata.Labels,
				OwnerReferences: []object.OwnerReference{{
					Kind:       "Deployment",
					Name:       deployment.Metadata.Name,
					UID:        "",
					Controller: false,
				}},
			},
			Spec: object.ReplicaSetSpec{
				Replicas: deployment.Spec.Replicas,
				Template: deployment.Spec.Template,
			},
			Status: object.ReplicaSetStatus{Replicas: 0},
		}
		dc.dm2rs.Put(res.Key, rsKeyNew)
		err = client.Put(dc.apiServerBase+rsKeyNew, rs)
		if err != nil {
			klog.Errorf("Error send new rs to etcd\n")
		}
	} else if res.IsModify {
		update := func() {
			rsKeyNew := "/registry/rs/default/" + deployment.Metadata.Name + uuid.New().String()
			rsKeyOld, _ := dc.dm2rs.Get(res.Key)
			fmt.Println(deployment)
			surge := *deployment.Spec.Strategy.RollingUpdate.MaxSurge
			replicas := deployment.Spec.Replicas
			vs, ok := dc.replicasetMap.Get(rsKeyOld)
			var decreaseOldDone, increaseNewDone bool
			increaseNewDone = false
			var rsOld object.ReplicaSet
			if !ok {
				decreaseOldDone = true
			} else {
				rsOld = vs.Replicaset
				decreaseOldDone = rsOld.Spec.Replicas <= 0
			}
			rsNew := object.ReplicaSet{
				ObjectMeta: object.ObjectMeta{
					Name:   deployment.Metadata.Name,
					Labels: deployment.Metadata.Labels,
					OwnerReferences: []object.OwnerReference{{
						Kind:       "Deployment",
						Name:       deployment.Metadata.Name,
						UID:        "",
						Controller: false,
					}},
				},
				Spec: object.ReplicaSetSpec{
					Replicas: func() int32 {
						if surge < replicas {
							return surge
						} else {
							return replicas
						}
					}(),
					Template: deployment.Spec.Template,
				},
				Status: object.ReplicaSetStatus{Replicas: 0},
			}
			err = client.Put(dc.apiServerBase+rsKeyNew, rsNew)
			if err != nil {
				klog.Errorf("%s\n", err.Error())
			}
			increaseNewDone = rsNew.Spec.Replicas == replicas
			decreaseOldDone = rsOld.Spec.Replicas == 0

			for true {
				if !increaseNewDone {
					stash := rsNew.Spec.Replicas
					rsNew.Spec.Replicas += 1
					if rsNew.Spec.Replicas > replicas {
						increaseNewDone = true
					} else {
						err = client.Put(dc.apiServerBase+rsKeyNew, rsNew)
						if err != nil {
							rsNew.Spec.Replicas = stash
							klog.Errorf("%s\n", err.Error())
							goto LoopErr
						}
					}
				}
				if !decreaseOldDone {
					stash := rsOld.Spec.Replicas
					rsOld.Spec.Replicas -= 1
					if rsOld.Spec.Replicas > 0 {
						err = client.Put(dc.apiServerBase+rsKeyOld, rsOld)
						if err != nil {
							klog.Errorf("%s\n", err.Error())
							rsOld.Spec.Replicas = stash
							goto LoopErr
						}
					} else {
						err = client.Del(dc.apiServerBase + rsKeyOld)
						decreaseOldDone = true
						if err != nil {
							klog.Errorf("%s\n", err.Error())
							rsOld.Spec.Replicas = stash
							goto LoopErr
						}
					}
				}
				if decreaseOldDone && increaseNewDone {
					break
				}
			LoopErr:
				time.Sleep(15 * time.Second)
			}
			dc.dm2rs.Put(res.Key, rsKeyNew)
		}
		go update()
	}
}

func (dc *DeploymentController) watchReplicaset(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		rs := object.ReplicaSet{}
		err := json.Unmarshal(res.ValueBytes, &rs)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		if rs.Spec.Replicas == 0 {
			dc.replicasetMap.Del(res.Key)
		} else {
			vrs := object.VersionedReplicaset{
				Version:    res.ResourceVersion,
				Replicaset: rs,
			}
			dc.replicasetMap.Put(res.Key, vrs)
		}
		break
	case etcdstore.DELETE:
		dc.replicasetMap.Del(res.Key)
		break
	}
}

func (dc *DeploymentController) deleteDeployment(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.DELETE {
		return
	}
	rsKey, _ := dc.dm2rs.Get(res.Key)
	dc.deploymentMap.Del(res.Key)
	dc.dm2rs.Del(res.Key)
	err := client.Del(dc.apiServerBase + rsKey)
	if err != nil {
		klog.Errorf("Error del rs %s. Err : %s\n", dc.apiServerBase+rsKey, err.Error())
		return
	}
}
