package replicaset

import (
	"context"
	"github.com/emicklei/go-restful"
	"minik8s/k8s.io/client/rest"
	workqueue "minik8s/k8s.io/client/util"
	"minik8s/object"
	"minik8s/pkg/controller"
)

type ReplicaSetController struct {
	// client connect to api server
	kubeClient rest.Interface
	kubeServer *restful.WebService

	// connect to pod
	podControl controller.PodControlInterface

	syncHandler func(ctx context.Context, rsKey string) error

	queue workqueue.Interface
}

func NewReplicaSetController() *ReplicaSetController {

	return &ReplicaSetController{
		kubeServer: new(restful.WebService),
	}
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run(ctx context.Context) {

	go rsc.worker(ctx)
	<-ctx.Done()
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
func (rsc *ReplicaSetController) worker(ctx context.Context) {
	for {
		key, _ := rsc.queue.Get()
		rsc.syncHandler(ctx, key.(string))
	}
}

func (rsc *ReplicaSetController) register() {

	rsc.kubeServer.Route(rsc.kubeServer.GET("/addRS").To(rsc.addRS))
	restful.Add(rsc.kubeServer)
}

// TODO: handle add replica set
func (rsc *ReplicaSetController) addRS(req *restful.Request, resp *restful.Response) {
	rs := &object.ReplicaSet{}
	rsc.enqueueRS(rs)
}

// When a pod is created, enqueue its replica set
func (rsc *ReplicaSetController) addPod(obj interface{}) {

}

// TODO: add replicaset to queue
func (rsc *ReplicaSetController) enqueueRS(rs *object.ReplicaSet) {
	key := ""
	rsc.queue.Add(key)
}

// syncReplicaSet will sync the ReplicaSet with the given key if it has had its expectations fulfilled,
// meaning it did not expect to see any more of its pods created or deleted. This function is not meant to be
// invoked concurrently with the same key.
func (rsc *ReplicaSetController) syncReplicaSet(ctx context.Context, key string) error {
	return nil
}

// manageReplicas checks and updates replicas for the given ReplicaSet.
// It will requeue the replica set in case of an error while creating/deleting pods.
func (rsc *ReplicaSetController) manageReplicas(ctx context.Context, filteredPods []*object.Pod, rs *object.ReplicaSet) error {
	// 调用 apiserver 的接口进行创建
	return nil
}
