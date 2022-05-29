package replicaset

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/controller"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrent_map "minik8s/util/map"
	"minik8s/util/queue"
	"sync"
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

func NewReplicaSetController(ctx context.Context, controllerCtx util.ControllerContext) *ReplicaSetController {
	restClient := client.RESTClient{
		Base: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}

	cp := concurrent_map.NewConcurrentMap()

	rsc := &ReplicaSetController{
		ls:     controllerCtx.Ls,
		cp:     cp,
		Client: restClient,
	}
	return rsc
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run(ctx context.Context) {
	klog.Debugf("[ReplicaSetController]start running\n")
	go rsc.register()
	go rsc.worker(ctx)
	select {}
}

func (rsc *ReplicaSetController) register() {
	watchAdd := func(rsc *ReplicaSetController) {
		err := rsc.ls.Watch(config.RSConfigPrefix, rsc.addRS, rsc.stopChannel)
		if err != nil {
			fmt.Printf("[ReplicaSetController] ListWatch init fail...")
		}
	}

	watchPod := func(rsc *ReplicaSetController) {
		err := rsc.ls.Watch("/registry/pod/default", rsc.podOperation, rsc.stopChannel)
		if err != nil {
			fmt.Printf("[ReplicaSetController] ListWatch init fail...")
		}
	}

	klog.Debugf("success register\n")

	go watchAdd(rsc)
	go watchPod(rsc)

}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
func (rsc *ReplicaSetController) worker(ctx context.Context) {
	var m sync.Mutex
	for {
		if !rsc.queue.Empty() {
			key := rsc.queue.Front()
			rsc.queue.Dequeue()
			m.Lock()
			rsc.syncReplicaSet(ctx, key.(string))
			m.Unlock()
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (rsc *ReplicaSetController) addRS(res etcdstore.WatchRes) {
	// do not react to delete, delete is mocked by put
	if res.ResType == etcdstore.DELETE {
		return
	}
	rs := &object.ReplicaSet{}
	err := json.Unmarshal(res.ValueBytes, rs)
	if err != nil {
		fmt.Printf("addRS bad message\n")
		return
	}

	fmt.Printf("[addRS] message receive...\n")

	// encode object to key
	key := getKey(rs)
	rsc.cp.Put(key, rs)
	// enqueue key
	rsc.queue.Enqueue(key)
}

func (rsc *ReplicaSetController) podOperation(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		return
	}
	// check ownership
	pod := &object.Pod{}
	fmt.Printf("[podOperation] messgae:%v\n", len(res.ValueBytes))
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("[podOperation] bad message,unmarshal fail\n")
		return
	}

	if pod.Status.Phase == "" {
		return
	}

	fmt.Printf("[podOperation] pod message:%v\n", pod)

	isOwned, name, _ := client.OwnByRs(pod)
	if isOwned {
		rs, err := client.GetRuntimeRS(rsc.ls, name)
		//fmt.Printf("[podOperation] rs:%v owns:%v\n", rs.NodeName, pod.NodeName)
		if err == nil {
			// encode object to key
			key := getKey(rs)
			rsc.cp.Put(key, rs)
			// enqueue key
			rsc.queue.Enqueue(key)
		}
	}
}

func (rsc *ReplicaSetController) syncReplicaSet(ctx context.Context, key string) error {
	// get expected replica set
	rs, _ := rsc.cp.Get(key).(*object.ReplicaSet)
	// get all actual pods of the rs
	allPods, _ := client.GetRSPods(rsc.ls, rs.Name, rs.UID)
	// filter all inactive pods
	activePods := controller.FilterActivePods(allPods)
	fmt.Printf("[syncReplicaSet] active pods of rs %v:%v\n", rs.Name, len(activePods))
	if len(activePods) == int(rs.Spec.Replicas) {
		return nil
	}
	// manage pods
	err := rsc.manageReplicas(ctx, activePods, rs)
	// calculate new status
	newStatus := calculateStatus(rs, activePods)
	// update status
	err = putReplicaSet(ctx, &rsc.Client, rs, newStatus)
	return err
}

func (rsc *ReplicaSetController) manageReplicas(ctx context.Context, filteredPods []*object.Pod, rs *object.ReplicaSet) error {
	// make diff for current pods and expected number
	diff := len(filteredPods) - int(rs.Spec.Replicas)

	fmt.Printf("[manageReplicas] active:%v expected:%v diff:%v\n", len(filteredPods), int(rs.Spec.Replicas), diff)
	fmt.Printf("[manageReplicas] active:")
	for _, p := range filteredPods {
		fmt.Printf("%s  ", p.Name)
	}
	fmt.Printf("\n")

	if diff < 0 {
		diff *= -1
		// create pods
		for i := 0; i < diff; i++ {
			fmt.Printf("[manageReplicas] create pod\n")
			err := rsc.Client.CreateRSPod(ctx, rs)
			if err != nil {
				klog.Errorf("create pod fail\n")
			}
		}

	} else if diff > 0 {
		// delete pods
		podsToDelete := getPodsToDelete(filteredPods, diff)
		fmt.Printf("[manageReplicas] del pods number:%v\n", len(podsToDelete))

		for _, pod := range podsToDelete {
			err := rsc.Client.DeleteConfigPod(pod.Name)
			if err != nil {
				klog.Errorf("delete pod config fail Name:%s uid:%s\n", pod.Name, pod.UID)
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

// choose pods to be deleted
// simple policy
func getPodsToDelete(filteredPods []*object.Pod, diff int) []*object.Pod {
	return filteredPods[:diff]
}

func getKey(rs *object.ReplicaSet) string {
	return rs.Name + rs.UID
}

func putReplicaSet(ctx context.Context, c *client.RESTClient, rs *object.ReplicaSet, newStatus object.ReplicaSetStatus) error {
	rs.Status = newStatus
	var err error

	if rs.Spec.Replicas == 0 {
		// do real deletion
		err = c.DeleteRS(rs.Name)
	} else {
		err = c.PutWrap("/registry/rs/default/"+rs.Name, rs)
	}

	return err
}
