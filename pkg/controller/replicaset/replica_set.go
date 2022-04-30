package replicaset

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/controller"
	"minik8s/pkg/klog"
	"minik8s/pkg/queue"
	"time"

	"minik8s/pkg/messaging"
)

type ReplicaSetController struct {
	// watcher
	Subscriber   *messaging.Subscriber
	ExchangeName string
	stopCh       <-chan struct{}

	// working queue
	queue queue.ConcurrentQueue

	Client client.RESTClient
}

func NewReplicaSetController(msgConfig messaging.QConfig, clientConfig client.Config) *ReplicaSetController {
	subscriber, _ := messaging.NewSubscriber(&msgConfig)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	rsc := &ReplicaSetController{
		Subscriber: subscriber,
		Client:     restClient,
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
	exchangeName, _, err := rsc.Client.WatchRegister("rs", "", true)
	if err != nil {
		klog.Errorf("register watchNewPod fail\n")
	}
	err = rsc.Subscriber.Subscribe(*exchangeName, rsc.addRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("subscribe watchNewPod fail\n")
	}

	exchangeName, _, err = rsc.Client.WatchRegister("rs", "", true)
	if err != nil {
		klog.Errorf("register watchNewPod fail\n")
	}
	err = rsc.Subscriber.Subscribe(*exchangeName, rsc.deleteRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("subscribe watchNewPod fail\n")
	}

	exchangeName, _, err = rsc.Client.WatchRegister("rs", "", true)
	if err != nil {
		klog.Errorf("register watchNewPod fail\n")
	}
	err = rsc.Subscriber.Subscribe(*exchangeName, rsc.updateRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("subscribe watchNewPod fail\n")
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

func (rsc *ReplicaSetController) addRS(d amqp.Delivery) {
	rs := &object.ReplicaSet{}
	err := json.Unmarshal(d.Body, rs)
	if err != nil {
		klog.Warnf("addRS bad message\n")
	}
	// encode object to key
	key := getKey(rs)
	// enqueue key
	rsc.queue.Enqueue(key)
}

func (rsc *ReplicaSetController) updateRS(d amqp.Delivery) {

}

func (rsc *ReplicaSetController) deleteRS(d amqp.Delivery) {

}

func (rsc *ReplicaSetController) syncReplicaSet(ctx context.Context, key string) error {
	// get name of key
	name := key
	// get all replica sets of the name
	rs, _ := rsc.Client.GetRS(name)
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

// TODO: key is the resource name
func getKey(rs *object.ReplicaSet) string {
	return rs.Name
}

func updateReplicaSetStatus(ctx context.Context, c *client.RESTClient, rs *object.ReplicaSet, newStatus object.ReplicaSetStatus) (*object.ReplicaSet, error) {
	var updatedRS *object.ReplicaSet
	updatedRS, _ = c.UpdateRSStatus(ctx, rs)
	return updatedRS, nil
}
