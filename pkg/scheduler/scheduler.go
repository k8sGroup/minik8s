package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/listerwatcher"
	"minik8s/util/queue"
	"time"
)

type Scheduler struct {
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	queue       queue.ConcurrentQueue
	Client      client.RESTClient
}

func NewScheduler(lsConfig *listerwatcher.Config, clientConfig client.Config) *Scheduler {
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Printf("[Scheduler] list watch start fail...")
	}

	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}

	rsc := &Scheduler{
		ls:     ls,
		Client: restClient,
	}
	return rsc
}

// Run begins watching and syncing.
func (sched *Scheduler) Run(ctx context.Context) {
	fmt.Printf("[Scheduler]start running\n")
	go sched.register()
	go sched.worker(ctx)
	select {}
}

func (sched *Scheduler) register() {
	err := sched.ls.Watch("/registry/pod/default", sched.watchNewPod, sched.stopChannel)
	if err != nil {
		fmt.Printf("[Scheduler] ListWatch init fail...\n")
	}
}

func (sched *Scheduler) worker(ctx context.Context) {
	fmt.Printf("[worker] Starting...\n")
	for {
		if !sched.queue.Empty() {
			podPtr := sched.queue.Front()
			sched.queue.Dequeue()
			sched.schedulePod(ctx, podPtr.(*object.Pod))
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (sched *Scheduler) schedulePod(ctx context.Context, pod *object.Pod) error {
	fmt.Printf("[schedulePod] Begin scheduling\n")
	nodes, _ := client.GetNodes(sched.ls)
	// select a host for the pod
	nodeName, _ := selectHost(nodes)
	fmt.Printf("[schedulePod]assign pod to node:%s\n", nodeName)
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
	var idx int
	if num == 1 {
		idx = 0
	} else {
		rand.Seed(time.Now().UnixNano())
		idx = rand.Intn(num - 1)
	}
	return nodes[idx].Name, nil
}

// watch the change of new pods
func (sched *Scheduler) watchNewPod(res etcdstore.WatchRes) {
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("watchNewPod bad message\n")
		return
	}
	// check whether scheduled
	if pod.Spec.NodeName != "" {
		return
	}

	fmt.Printf("[watchNewPod] new message from watcher...\n")
	sched.queue.Enqueue(pod)
}
