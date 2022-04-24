package replicaset

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"minik8s/object"
	"minik8s/pkg/klog"
	"minik8s/pkg/queue"

	"minik8s/pkg/messaging"
)

type ReplicaSetController struct {
	// watch
	Subscriber   *messaging.Subscriber
	ExchangeName string
	stopCh       <-chan struct{}

	queue queue.ConcurrentQueue
}

func NewReplicaSetController(config messaging.QConfig) *ReplicaSetController {
	subscriber, _ := messaging.NewSubscriber(config)
	exchangeName := "ReplicaSetController"
	rsc := &ReplicaSetController{
		Subscriber:   subscriber,
		ExchangeName: exchangeName,
	}
	return rsc
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run(ctx context.Context) {
	rsc.register()
	go rsc.worker(ctx)
	<-ctx.Done()
}

func (rsc *ReplicaSetController) register() {
	err := rsc.Subscriber.Subscribe(rsc.ExchangeName+"."+"addRS", rsc.addRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("register addRS fail")
	}
	err = rsc.Subscriber.Subscribe(rsc.ExchangeName+"."+"updateRS", rsc.updateRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("register updateRS fail")
	}
	err = rsc.Subscriber.Subscribe(rsc.ExchangeName+"."+"deleteRS", rsc.deleteRS, rsc.stopCh)
	if err != nil {
		klog.Errorf("register deleteRS fail")
	}
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
func (rsc *ReplicaSetController) worker(ctx context.Context) {
	for {
		key := rsc.queue.Front()
		rsc.syncReplicaSet(ctx, key)
	}
}

func (rsc *ReplicaSetController) addRS(d amqp.Delivery) {
	rs := &object.ReplicaSet{}
	err := json.Unmarshal(d.Body, rs)
	if err != nil {
		klog.Warnf("addRS bad message")
	}
	// store key and value
	key := ""

	// enqueue key
	rsc.queue.Enqueue(key)
}

func (rsc *ReplicaSetController) updateRS(d amqp.Delivery) {

}

func (rsc *ReplicaSetController) deleteRS(d amqp.Delivery) {

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
