package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/listerwatcher"
	"minik8s/util/queue"
	"sync"
	"time"
)

var globalCount int
var lock sync.Mutex

const (
	SelectRandom     string = "1"
	SelectRoundRobin string = "2"
	SelectAffinity   string = "3"
)

type Scheduler struct {
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	queue       queue.ConcurrentQueue
	selectType  string
	Client      client.RESTClient
}

func NewScheduler(lsConfig *listerwatcher.Config, clientConfig client.Config, selectType string) *Scheduler {
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("[Scheduler] list watch start fail...")
	}

	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}

	rsc := &Scheduler{
		ls:     ls,
		Client: restClient,
	}
	rsc.stopChannel = make(chan struct{})
	rsc.selectType = selectType
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
	err := sched.ls.Watch(config.PodConfigPREFIX, sched.watchNewPod, sched.stopChannel)
	if err != nil {
		fmt.Println(err)
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
	nodes, err := sched.getNodes()
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	// select a host for the pod
	var nodeName string
	switch sched.selectType {
	case SelectRandom:
		nodeName = selectHostRandom(nodes)
		break
	case SelectRoundRobin:
		nodeName = selectHostRoundRobin(nodes)
		break
	case SelectAffinity:
		nodeName = selectHostWithAffinity(nodes, pod.Labels)
		break
	}
	fmt.Printf("the nodeName choice is:%s\n", nodeName)
	fmt.Printf("[schedulePod]assign pod to node:%s\n", nodeName)
	// modify pod host
	pod.Spec.NodeName = nodeName
	// update pod to api server
	err = sched.Client.UpdateConfigPod(pod)
	return err
}
func (sched *Scheduler) getNodes() ([]object.Node, error) {
	raw, err := sched.ls.List(config.NODE_PREFIX)
	if err != nil {
		return nil, err
	}
	var res []object.Node
	if len(raw) == 0 {
		return res, nil
	}
	for _, rawPair := range raw {
		node := &object.Node{}
		err = json.Unmarshal(rawPair.ValueBytes, node)
		res = append(res, *node)
	}
	return res, nil
}

// select a node as host
// select police
func selectHostRandom(nodes []object.Node) string {
	if len(nodes) == 0 {
		return ""
	}
	num := len(nodes)
	fmt.Printf("there are %d nodes in totle", num)
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(num)
	return nodes[idx].MetaData.Name
}
func selectHostRoundRobin(nodes []object.Node) string {
	lock.Lock()
	defer lock.Unlock()
	num := len(nodes)
	if num == 0 {
		return ""
	}
	idx := globalCount % num
	globalCount++
	return nodes[idx].MetaData.Name
}
func match(map1 map[string]string, map2 map[string]string) bool {
	for k, v := range map1 {
		v2, ok := map2[k]
		if !ok {
			continue
		} else {
			if v == v2 {
				return true
			}
			continue
		}
	}
	return false
}

//使用labels进行挑选
func selectHostWithAffinity(nodes []object.Node, labels map[string]string) string {
	if labels == nil {
		return selectHostRandom(nodes)
	}
	for _, node := range nodes {
		if node.MetaData.Labels == nil {
			continue
		}
		if match(labels, node.MetaData.Labels) {
			return node.MetaData.Name
		} else {
			continue
		}
	}
	return selectHostRandom(nodes)
}

// watch the change of new pods
func (sched *Scheduler) watchNewPod(res etcdstore.WatchRes) {
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		fmt.Printf("watchNewPod bad message pod:%+v\n", pod)
		return
	}

	if pod.Spec.NodeName != "" {
		return
	}

	// check whether scheduled
	fmt.Printf("watch new Config Pod with name:%s\n", pod.Name)

	fmt.Printf("[watchNewPod] new message from watcher...\n")
	sched.queue.Enqueue(pod)
}
