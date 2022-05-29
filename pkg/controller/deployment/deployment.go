package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"path"
	"time"
)

type RsPodStatus struct {
	Actual int32 `json:"actual" yaml:"actual"`
	Expect int32 `json:"expect" yaml:"expect"`
}

type DeploymentController struct {
	ls             *listerwatcher.ListerWatcher
	deploymentMap  *concurrentmap.ConcurrentMapTrait[string, object.VersionedDeployment]
	replicasetMap  *concurrentmap.ConcurrentMapTrait[string, object.ReplicaSet]
	dm2rs          *concurrentmap.ConcurrentMapTrait[string, string] // dm2rs mapping from Deployment key in etcdstore to Replicaset key in etcdstore
	resyncInterval time.Duration
	stopChannel    chan struct{}
	apiServerBase  string
}

func NewDeploymentController(ctx context.Context, controllerCtx util.ControllerContext) *DeploymentController {
	dc := &DeploymentController{
		ls:             controllerCtx.Ls,
		deploymentMap:  concurrentmap.NewConcurrentMapTrait[string, object.VersionedDeployment](),
		replicasetMap:  concurrentmap.NewConcurrentMapTrait[string, object.ReplicaSet](),
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
	//registerWatchReplicaset := func() {
	//	for {
	//		err := dc.ls.Watch(config.RSConfigPrefix, dc.watchReplicaset, dc.stopChannel)
	//		if err != nil {
	//			klog.Errorf("Error watching %s\n", config.RSConfigPrefix)
	//		} else {
	//			return
	//		}
	//		time.Sleep(5 * time.Second)
	//	}
	//}

	registerAddDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment/default", dc.putDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment : %s\n", err.Error())
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerDeleteDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment/default", dc.deleteDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment : %s\n", err.Error())
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerResyncLoop := func() {
		for {
			//{
			//	resList, err := dc.ls.List(config.RSConfigPrefix)
			//	if err != nil {
			//		klog.Errorf("Error synchronizing!\n")
			//		goto failed
			//	}
			//	newMap := make(map[string]object.VersionedReplicaset)
			//	for _, res := range resList {
			//		versionedRS := object.VersionedReplicaset{Version: res.ResourceVersion}
			//		_ = json.Unmarshal(res.ValueBytes, &versionedRS.Replicaset)
			//		newMap[res.Key] = versionedRS
			//	}
			//	dc.replicasetMap.UpdateAll(newMap, object.SelectNewerReplicaset)
			//}
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
		rsUidNew := uuid.New().String()
		rsNameNew := deployment.Metadata.Name + rsUidNew
		rsKeyNew := path.Join(config.RSConfigPrefix, rsNameNew)
		rs := object.ReplicaSet{
			ObjectMeta: object.ObjectMeta{
				Name:   rsNameNew,
				Labels: deployment.Metadata.Labels,
				UID:    rsUidNew,
				OwnerReferences: []object.OwnerReference{{
					Kind:       "Deployment",
					Name:       deployment.Metadata.Name,
					UID:        deployment.Metadata.UID,
					Controller: false,
				}},
			},
			Spec: object.ReplicaSetSpec{
				Replicas: deployment.Spec.Replicas,
				Template: deployment.Spec.Template,
			},
		}
		dc.dm2rs.Put(res.Key, rsKeyNew)

		err = client.Put(dc.apiServerBase+rsKeyNew, rs)
		if err != nil {
			klog.Errorf("Error send new rs to etcd\n")
		}
		dc.replicasetMap.Put(rsKeyNew, rs)
	} else if res.IsModify {
		update := func() {
			fmt.Println("go update")
			rsUidNew := uuid.New().String()
			rsNameNew := deployment.Metadata.Name + rsUidNew
			//rsUidOld := ""
			rsNameOld := ""
			rsKeyNew := path.Join(config.RSConfigPrefix, rsNameNew)
			rsKeyOld, _ := dc.dm2rs.Get(res.Key)
			fmt.Println(deployment)
			surge := *deployment.Spec.Strategy.RollingUpdate.MaxSurge
			replicas := deployment.Spec.Replicas
			rsOld, isOldRSExist := dc.replicasetMap.Get(rsKeyOld)
			var decreaseOldDone, increaseNewDone bool
			increaseNewDone = false
			if !isOldRSExist {
				// old replicaset doesn't exist
				decreaseOldDone = true
			} else {
				rsNameOld = rsOld.Name
				decreaseOldDone = rsOld.Spec.Replicas <= 0
			}

			// clear old replicaset's owner
			if isOldRSExist && !decreaseOldDone {
				rsOld.OwnerReferences = []object.OwnerReference{}
				err = client.Put(dc.apiServerBase+rsKeyOld, rsOld)
				if err != nil {
					klog.Errorf("%s\n", err.Error())
				}
				dc.replicasetMap.Put(rsKeyOld, rsOld)
			}

			// create new replicaset and set its owner
			rsNew := object.ReplicaSet{
				ObjectMeta: object.ObjectMeta{
					Name:   rsNameNew,
					UID:    rsUidNew,
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
			}

			//dc.dm2rs.Put(res.Key, rsKeyNew)

			for true {
				fmt.Printf("[delta loop]\n")
				fmt.Printf("[new rs] %s - %d\n", rsNameNew, rsNew.Spec.Replicas)
				fmt.Printf("[old rs] %s - %d\n", rsNameOld, rsOld.Spec.Replicas)

				if !increaseNewDone {
					stash := rsNew.Spec.Replicas
					fmt.Printf("[send new rs] %s - %d\n", rsNameNew, rsNew.Spec.Replicas)
					err = client.Put(dc.apiServerBase+rsKeyNew, rsNew)
					if err != nil {
						rsNew.Spec.Replicas = stash
						fmt.Printf("[error] send new rs %s %s\n", rsNameNew, err.Error())
						goto LoopErr
					}
					dc.replicasetMap.Put(rsKeyNew, rsNew)
					rsNew.Spec.Replicas += 1
					if rsNew.Spec.Replicas > replicas {
						increaseNewDone = true
					}
				}

				time.Sleep(7 * time.Second)

				if !decreaseOldDone {
					stash := rsOld.Spec.Replicas
					rsOld.Spec.Replicas -= 1
					fmt.Printf("[send old rs] %s - %d\n", rsNameOld, rsOld.Spec.Replicas)
					if rsOld.Spec.Replicas > 0 {
						err = client.Put(dc.apiServerBase+rsKeyOld, rsOld)
						if err != nil {
							fmt.Printf("[error] send old rs %s %s\n", rsNameOld, err.Error())
							rsOld.Spec.Replicas = stash
							goto LoopErr
						}
						dc.replicasetMap.Put(rsKeyOld, rsOld)
					} else if rsOld.Spec.Replicas == 0 {
						err = client.Put(dc.apiServerBase+rsKeyOld, rsOld)
						if err != nil {
							fmt.Printf("[error] send old rs %s %s\n", rsNameOld, err.Error())
							rsOld.Spec.Replicas = stash
							goto LoopErr
						}
						dc.replicasetMap.Put(rsKeyOld, rsOld)
						decreaseOldDone = true
					} else {
						decreaseOldDone = true
					}
				}

				if decreaseOldDone && increaseNewDone {
					fmt.Printf("[old rs decreased] %s\n", rsNameOld)
					fmt.Printf("[new rs increased] %s\n", rsNameNew)
					break
				}

				//for true {
				//	var status RsPodStatus
				//	time.Sleep(time.Second)
				//	if !syncNew && !increaseNewDone {
				//		data, err := client.GetWithParams(dc.apiServerBase+config.RS_POD, map[string]string{"rsName": rsNameNew, "uid": rsUidNew})
				//		if err != nil {
				//			continue
				//		}
				//		err = json.Unmarshal(data, &status)
				//		if err != nil {
				//			continue
				//		}
				//		fmt.Printf("[new rs status] rs %s actual %d expected %d\n", rsNameNew, status.Actual, status.Expect)
				//		if status.Actual == status.Expect {
				//			syncNew = true
				//		}
				//	}
				//	if !syncOld && !decreaseOldDone {
				//		data, err := client.GetWithParams(dc.apiServerBase+config.RS_POD, map[string]string{"rsName": rsNameOld, "uid": rsUidOld})
				//		if err != nil {
				//			continue
				//		}
				//		err = json.Unmarshal(data, &status)
				//		if err != nil {
				//			continue
				//		}
				//		fmt.Printf("[old rs status] rs %s actual %d expected %d\n", rsNameOld, status.Actual, status.Expect)
				//		if status.Actual == status.Expect {
				//			syncOld = true
				//		}
				//	}
				//	if syncNew && syncOld {
				//		fmt.Println("[status check] actual==expected, next loop")
				//		goto NextLoop
				//	}
				//}
			LoopErr:
				time.Sleep(7 * time.Second)
			}
			dc.dm2rs.Put(res.Key, rsKeyNew)
		}
		go update()
	}
}

//func (dc *DeploymentController) watchReplicaset(res etcdstore.WatchRes) {
//	switch res.ResType {
//	case etcdstore.PUT:
//		rs := object.ReplicaSet{}
//		err := json.Unmarshal(res.ValueBytes, &rs)
//		if err != nil {
//			klog.Errorf("%s\n", err.Error())
//			return
//		}
//		if rs.Spec.Replicas == 0 {
//			dc.replicasetMap.Del(res.Key)
//		} else {
//			vrs := object.VersionedReplicaset{
//				Version:    res.ResourceVersion,
//				Replicaset: rs,
//			}
//			dc.replicasetMap.Put(res.Key, vrs)
//		}
//		break
//	case etcdstore.DELETE:
//		dc.replicasetMap.Del(res.Key)
//		break
//	}
//}

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
