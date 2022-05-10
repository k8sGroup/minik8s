package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"time"
)

type versionedDeployment struct {
	version    int64
	deployment object.Deployment
}

func selectNewerDeployment(d1 versionedDeployment, d2 versionedDeployment) versionedDeployment {
	if d1.version > d2.version {
		return d1
	} else {
		return d2
	}
}

type versionedReplicaset struct {
	version    int64
	replicaset object.ReplicaSet
}

func selectNewerReplicaset(rs1 versionedReplicaset, rs2 versionedReplicaset) versionedReplicaset {
	if rs1.version > rs2.version {
		return rs1
	} else {
		return rs2
	}
}

type DeploymentController struct {
	ls             *listerwatcher.ListerWatcher
	deploymentMap  *concurrentmap.ConcurrentMapTrait[string, versionedDeployment]
	replicasetMap  *concurrentmap.ConcurrentMapTrait[string, versionedReplicaset]
	resyncInterval time.Duration
	stopChannel    chan struct{}
	apiServerBase  string
}

func NewDeploymentController(ctx context.Context, controllerCtx util.ControllerContext) *DeploymentController {
	dc := &DeploymentController{
		ls:            controllerCtx.Ls,
		deploymentMap: concurrentmap.NewConcurrentMapTrait[string, versionedDeployment](),
		replicasetMap: concurrentmap.NewConcurrentMapTrait[string, versionedReplicaset](),
		stopChannel:   make(chan struct{}),
		apiServerBase: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}
	if dc.apiServerBase == "" {
		klog.Fatalf("uninitialized apiserver base!\n")
	}
	return dc
}

func (dc *DeploymentController) Run(ctx context.Context) {
	klog.Debugf("[DeploymentController] running...\n")
	// TODO
	<-ctx.Done()
	close(dc.stopChannel)
}

func (dc *DeploymentController) register() {
	registerAddDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment", dc.putDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment\n")
			}
			time.Sleep(5 * time.Second)
		}
	}
	registerDeleteDeployment := func() {
		for {
			err := dc.ls.Watch("/registry/deployment", dc.deleteDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment\n")
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerDeploymentResyncLoop := func() {
		//TODO handle the event when the list responses mismatch with local cache
		for {
			{
				resList, err := dc.ls.List("/registry/rs/default")
				if err != nil {
					klog.Errorf("Error synchronizing!\n")
					continue
				}
				newMap := make(map[string]versionedReplicaset)
				for _, res := range resList {
					versionedRS := versionedReplicaset{version: res.ResourceVersion}
					_ = json.Unmarshal(res.ValueBytes, &versionedRS.replicaset)
					newMap[res.Key] = versionedRS
				}
				dc.replicasetMap.UpdateAll(newMap, selectNewerReplicaset)
			}
			{
				resList, err := dc.ls.List("/registry/deployment/default")
				if err != nil {
					klog.Errorf("Error synchronizing!\n")
					continue
				}
				newMap := make(map[string]versionedDeployment)
				for _, res := range resList {
					versionedDM := versionedDeployment{version: res.ResourceVersion}
					_ = json.Unmarshal(res.ValueBytes, &versionedDM.deployment)
					newMap[res.Key] = versionedDM
				}
				dc.deploymentMap.ReplaceAll(newMap)
			}
			time.Sleep(dc.resyncInterval)
		}
	}

	go registerDeploymentResyncLoop()
	go registerAddDeployment()
	go registerDeleteDeployment()
}

func (dc *DeploymentController) putDeployment(res etcdstore.WatchRes) {
	key := res.Key
	// TODO 根据deployment的name确定replicaset的name
	var name string
	_, err := fmt.Sscanf(key, "/registry/deployment/default/%s", &name)
	if err != nil {
		klog.Errorf("Error parsing deployment key %s\n", key)
		return
	}
	deployment := object.Deployment{}
	err = json.Unmarshal(res.ValueBytes, &deployment)
	if err != nil {
		klog.Errorf("Error unmarshalling deployment json data\n")
		return
	}
	if res.IsCreate {
		// TODO : create a new replicaset and send it to etcd
		rsKey := "/registry/rs/default/" + name
		rs := object.ReplicaSet{
			ObjectMeta: deployment.Metadata,
			Spec: object.ReplicaSetSpec{
				Replicas: deployment.Spec.Replicas,
				Template: deployment.Spec.Template,
			},
			Status: object.ReplicaSetStatus{Replicas: 0},
		}
		err = client.Put(dc.apiServerBase+rsKey, rs)
		if err != nil {
			klog.Errorf("Error send new rs to etcd\n")
		}
	} else if res.IsModify {
		// TODO : update deployment

	}
}

func (dc *DeploymentController) deleteDeployment(res etcdstore.WatchRes) {

}

func (dc *DeploymentController) deltaDeployment(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		dc.putDeployment(res)
	case etcdstore.DELETE:
		dc.deleteDeployment(res)
	default:
		klog.Fatalf("Internal error!\n")
	}
}

func (dc *DeploymentController) syncReplicaset(res etcdstore.WatchRes) {

}