package replicaset

import (
	"context"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"minik8s/k8s.io/client/rest"
	workqueue "minik8s/k8s.io/client/util"
	"minik8s/object"
	"minik8s/pkg/controller"
)

type ReplicaSetController struct {
	// client connect to api server
	kubeClient rest.Interface
	// server provide service for api server
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
	body, _ := ioutil.ReadAll(req.Request.Body)
	_ = json.Unmarshal(body, rs)
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

func (rsc *ReplicaSetController) syncReplicaSet(ctx context.Context, key string) error {
	//namespace := "test"
	//name := "test"
	// get all replica sets of the namespace

	// get all pods of the namespace

	// filter all inactive pods

	// manage pods
	return nil
}

func (rsc *ReplicaSetController) manageReplicas(ctx context.Context, filteredPods []*object.Pod, rs *object.ReplicaSet) error {
	return nil
}
