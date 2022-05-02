package kubelet

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/kubelet/config"
	"minik8s/pkg/kubelet/podManager"
	"minik8s/pkg/kubelet/types"
	"minik8s/pkg/kubeproxy"
	"minik8s/pkg/kubeproxy/iptablesManager"
	"minik8s/pkg/listerwatcher"
)

type Kubelet struct {
	podManager *podManager.PodManager
	kubeproxy  *kubeproxy.Kubeproxy
	PodConfig  *config.PodConfig

	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
	Client      client.RESTClient
}

func NewKubelet(lsConfig *listerwatcher.Config, clientConfig client.Config) *Kubelet {
	kubelet := &Kubelet{}
	kubelet.podManager = podManager.NewPodManager()
	kubelet.kubeproxy = kubeproxy.NewKubeproxy()

	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	kubelet.Client = restClient

	// initialize list watch
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Printf("[NewKubelet] list watch start fail...")
	}
	kubelet.ls = ls

	// initialize pod config
	kubelet.PodConfig = config.NewPodConfig()

	return kubelet
}

func (kl *Kubelet) register() {
	err := kl.ls.Watch("/registry/pod/default", kl.watchPod, kl.stopChannel)
	if err != nil {
		fmt.Printf("[Kubelet] ListWatch init fail...")
	}
}

// Register TODO: node register to apiserver config
func (kl *Kubelet) registerNode() {
	meta := object.ObjectMeta{
		Name: "node1",
	}
	node := object.Node{
		ObjectMeta: meta,
	}
	err := kl.Client.RegisterNode(&node)
	if err != nil {
		fmt.Printf("[Kubelet] Register Node fail...")
	}
}

func (kl *Kubelet) Run() {
	kl.registerNode()
	go kl.register()

	updates := kl.PodConfig.GetUpdates()
	kl.syncLoop(updates, kl)
}

func (kl *Kubelet) syncLoop(updates <-chan types.PodUpdate, handler SyncHandler) {
	for {
		kl.syncLoopIteration(updates, handler)
	}
}

func (k *Kubelet) AddPod(pod *object.Pod) error {
	return k.podManager.AddPod(pod)
}
func (k *Kubelet) GetPodInfo(podName string) ([]byte, error) {
	return k.podManager.GetPodInfo(podName)
}
func (k *Kubelet) DeletePod(podName string) error {
	return k.podManager.DeletePod(podName)
}
func (k *Kubelet) AddPodPortMapping(podName string, podPort string, hostPort string) (iptablesManager.PortMapping, error) {
	p, err := k.podManager.GetPodSnapShoot(podName)
	if err != nil {
		return iptablesManager.PortMapping{}, err
	}
	return k.kubeproxy.AddPortMapping(p, podPort, hostPort)
}
func (k *Kubelet) RemovePortMapping(podName string, podPort string, hostPort string) error {
	p, err := k.podManager.GetPodSnapShoot(podName)
	if err != nil {
		return err
	}
	return k.kubeproxy.RemovePortMapping(p, podPort, hostPort)
}
func (k *Kubelet) GetPodMappingInfo() []iptablesManager.PortMapping {
	return k.kubeproxy.GetKubeproxySnapShoot().PortMappings
}

type SyncHandler interface {
	HandlePodAdditions(pods []*object.Pod)
	HandlePodUpdates(pods []*object.Pod)
	HandlePodRemoves(pods []*object.Pod)
	HandlePodReconcile(pods []*object.Pod)
	HandlePodSyncs(pods []*object.Pod)
	HandlePodCleanups() error
}

// TODO: channel pod type?
func (kl *Kubelet) syncLoopIteration(ch <-chan types.PodUpdate, handler SyncHandler) bool {
	select {
	case u, open := <-ch:
		if !open {
			fmt.Printf("Update channel is closed")
			return false
		}

		switch u.Op {
		case types.UPDATE:
			handler.HandlePodUpdates(u.Pods)
		case types.ADD:
			handler.HandlePodAdditions(u.Pods)
		case types.REMOVE:
			handler.HandlePodRemoves(u.Pods)
		case types.RECONCILE:
			handler.HandlePodReconcile(u.Pods)
		case types.DELETE:
			handler.HandlePodUpdates(u.Pods)
		}
	}
	return true
}

// TODO: check the message by node name. DO NOT handle pods not belong to this node
func (kl *Kubelet) watchPod(res etcdstore.WatchRes) {
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		klog.Warnf("watchNewPod bad message\n")
	}
	pods := []*object.Pod{pod}

	op := kl.getOpFromPod(pod)

	podUp := types.PodUpdate{
		Pods: pods,
		Op:   op,
	}
	kl.PodConfig.GetUpdates() <- podUp
}

func (kl *Kubelet) getOpFromPod(pod *object.Pod) types.PodOperation {
	op := types.ADD
	if pod.Status.Phase == object.PodFailed {
		op = types.DELETE
	}
	return op
}

func (kl *Kubelet) HandlePodAdditions(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] Prepare add pod:%+v\n", pod)
		err := kl.podManager.AddPod(pod)
		if err != nil {
			fmt.Printf("[Kubelet] Add pod fail...")
		}
	}
}

func (kl *Kubelet) HandlePodUpdates(pods []*object.Pod) {

}

func (kl *Kubelet) HandlePodRemoves(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] Prepare delete pod:%+v\n", pod)
		err := kl.podManager.DeletePod(pod.Name)
		if err != nil {
			fmt.Printf("[Kubelet] Add pod fail...")
		}
	}
}

func (kl *Kubelet) HandlePodReconcile(pods []*object.Pod) {

}

func (kl *Kubelet) HandlePodSyncs(pods []*object.Pod) {

}

func (kl *Kubelet) HandlePodCleanups() error {
	return nil
}
