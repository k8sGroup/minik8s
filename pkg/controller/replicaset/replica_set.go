package replicaset

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/cmd/kube-controller-manager/app"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/controller"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrent_map "minik8s/util/map"
	"minik8s/util/queue"
	"time"
)

type ReplicaSetController struct {
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}

	// working queue
	queue queue.ConcurrentQueue
	cp    *concurrent_map.ConcurrentMap

	Client client.RESTClient
}

func NewReplicaSetController(ctx context.Context, controllerCtx app.ControllerContext) *ReplicaSetController {
	restClient := client.RESTClient{
		Base: "http://" + controllerCtx.MasterIP,
	}

	cp := concurrent_map.NewConcurrentMap()

	rsc := &ReplicaSetController{
		ls:     controllerCtx.GetListerWatcher(),
		cp:     cp,
		Client: restClient,
	}
	return rsc
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run(ctx context.Context) {
	klog.Debugf("[ReplicaSetController]start running\n")
	rsc.register()
	go rsc.worker(ctx)
	<-ctx.Done()
}

func (rsc *ReplicaSetController) register() {
	err := rsc.ls.Watch("/registry/rs/default", rsc.addRS, rsc.stopChannel)
	if err != nil {
		fmt.Printf("[Scheduler] ListWatch init fail...")
	}

	err = rsc.ls.Watch("/registry/rs/default", rsc.deleteRS, rsc.stopChannel)
	if err != nil {
		fmt.Printf("[Scheduler] ListWatch init fail...")
	}

	klog.Debugf("success register\n")
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
func (rsc *ReplicaSetController) worker(ctx context.Context) {
	for {
		if !rsc.queue.Empty() {
			key := rsc.queue.Front()
			rsc.queue.Dequeue()
			rsc.syncReplicaSet(ctx, key.(string))
		} else {
			time.Sleep(time.Second)
		}
		klog.Infof("do some work\n")
	}
}

func (rsc *ReplicaSetController) addRS(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.PUT {
		return
	}

	fmt.Printf("[addRS] message receive...")

	rs := &object.ReplicaSet{}
	err := json.Unmarshal(res.ValueBytes, rs)
	if err != nil {
		klog.Warnf("addRS bad message\n")
	}
	// encode object to key
	key := getKey(rs)
	rsc.cp.Put(key, rs)
	// enqueue key
	rsc.queue.Enqueue(key)
}

func (rsc *ReplicaSetController) deleteRS(res etcdstore.WatchRes) {

	rs := &object.ReplicaSet{}
	err := json.Unmarshal(res.ValueBytes, rs)
	if err != nil {
		klog.Warnf("bad message\n")
	}

	// check whether the message is deletion
	if *rs.Spec.Replicas != 0 {
		return
	}

	fmt.Printf("[deleteRS] message receive...")

	// reset replicas to zero
	*rs.Spec.Replicas = 0

	// encode object to key
	key := getKey(rs)
	rsc.cp.Put(key, rs)
	// enqueue key
	rsc.queue.Enqueue(key)

}

func (rsc *ReplicaSetController) syncReplicaSet(ctx context.Context, key string) error {
	// get name of key
	name := key
	// get expected replica set
	rs, _ := rsc.cp.Get(key).(*object.ReplicaSet)
	// get all actual pods of the rs
	allPods, _ := rsc.Client.GetRSPods(name)
	// filter all inactive pods
	activePods := controller.FilterActivePods(allPods)
	// manage pods
	rsc.manageReplicas(ctx, activePods, rs)
	// calculate new status
	newStatus := calculateStatus(rs, activePods)
	// update status
	updateReplicaSetStatus(ctx, &rsc.Client, rs, newStatus)
	return nil
}

func (rsc *ReplicaSetController) manageReplicas(ctx context.Context, filteredPods []*object.Pod, rs *object.ReplicaSet) error {
	// make diff for current pods and expected number
	diff := len(filteredPods) - int(*(rs.Spec.Replicas))
	//key := getKey(rs)

	if diff < 0 {
		diff *= -1
		// create pods
		for i := 0; i < diff; i++ {
			err := rsc.Client.CreatePods(ctx, &rs.Spec.Template)
			if err != nil {
				klog.Errorf("create pod fail\n")
			}
		}

	} else if diff > 0 {
		// delete pods
		relatedPods, _ := rsc.getRelatedPods(rs)
		podsToDelete := getPodsToDelete(filteredPods, relatedPods, diff)
		for _, pod := range podsToDelete {
			err := rsc.Client.DeletePod(ctx, pod.Name)
			if err != nil {
				klog.Errorf("delete pod fail\n")
			}
		}
	}
	return nil
}

func calculateStatus(rs *object.ReplicaSet, filteredPods []*object.Pod) object.ReplicaSetStatus {
	newStatus := rs.Status
	newStatus.Replicas = int32(len(filteredPods))
	return newStatus
}

// ge related pods to replicaset
func (rsc *ReplicaSetController) getRelatedPods(rs *object.ReplicaSet) ([]*object.Pod, error) {
	var relatedPods []*object.Pod
	return relatedPods, nil
}

// choose pods to be deleted
// simple policy
func getPodsToDelete(filteredPods, relatedPods []*object.Pod, diff int) []*object.Pod {
	return filteredPods[:diff]
}

func getKey(rs *object.ReplicaSet) string {
	return rs.Name
}

func updateReplicaSetStatus(ctx context.Context, c *client.RESTClient, rs *object.ReplicaSet, newStatus object.ReplicaSetStatus) (*object.ReplicaSet, error) {
	var updatedRS *object.ReplicaSet
	updatedRS, _ = c.UpdateRSStatus(ctx, rs)
	return updatedRS, nil
}
