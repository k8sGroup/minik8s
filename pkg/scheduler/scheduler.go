package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"math/rand"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/klog"
	"minik8s/pkg/messaging"
	"minik8s/pkg/queue"
	"net/http"
	"net/url"
	"time"
)

type Scheduler struct {
	// watcher
	Subscriber *messaging.Subscriber
	stopCh     <-chan struct{}

	queue  queue.ConcurrentQueue
	Client client.RESTClient
}

func NewScheduler(msgConfig messaging.QConfig, clientConfig client.Config) *Scheduler {
	subscriber, _ := messaging.NewSubscriber(msgConfig)
	restClient := client.RESTClient{
		Client: &http.Client{},
		Base:   &url.URL{Host: "http://" + clientConfig.Host},
	}
	rsc := &Scheduler{
		Subscriber: subscriber,
		Client:     restClient,
	}
	return rsc
}

// Run begins watching and syncing.
func (sched *Scheduler) Run(ctx context.Context) {
	klog.Debugf("[ReplicaSetController]start running\n")
	sched.register()
	go sched.worker(ctx)
	<-ctx.Done()
}

func (sched *Scheduler) register() {
	exchangeName, _, err := sched.Client.WatchRegister("node", "", true)
	if err != nil {
		klog.Errorf("register watchNewPod fail\n")
	}
	err = sched.Subscriber.Subscribe(*exchangeName, sched.watchNewPod, sched.stopCh)
	if err != nil {
		klog.Errorf("subscribe watchNewPod fail\n")
	}
}

func (sched *Scheduler) worker(ctx context.Context) {
	for {
		if !sched.queue.Empty() {
			podPtr := sched.queue.Front()
			sched.queue.Dequeue()
			sched.schedulePod(ctx, podPtr.(*object.Pod))
		} else {
			time.Sleep(time.Second)
		}
		klog.Infof("do some work\n")
	}
}

func (sched *Scheduler) schedulePod(ctx context.Context, pod *object.Pod) error {
	nodes, _ := sched.Client.GetNodes()
	// select a host for the pod
	nodeName, _ := selectHost(nodes)
	// modify pod host
	pod.Spec.NodeName = nodeName
	// update pod to api server
	err := sched.Client.UpdatePods(ctx, pod)
	return err
}

// select a node as host
// TODO: change select policy
func selectHost(nodes []*object.Node) (string, error) {
	if len(nodes) == 0 {
		return "", errors.New("empty nodes")
	}
	num := len(nodes)
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(num - 1)
	return nodes[idx].Name, nil
}

// watch the change of new pods
func (sched *Scheduler) watchNewPod(d amqp.Delivery) {
	pod := &object.Pod{}
	err := json.Unmarshal(d.Body, pod)
	if err != nil {
		klog.Warnf("watchNewPod bad message\n")
	}
	sched.queue.Enqueue(pod)
}
