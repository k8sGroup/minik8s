package kubelet

import (
	"fmt"
	"minik8s/pkg/kubelet/module"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/kubelet/podManager"
	"minik8s/pkg/kubelet/types"
	"minik8s/pkg/kubeproxy"
	"minik8s/pkg/kubeproxy/iptablesManager"
)

type Kubelet struct {
	podManager *podManager.PodManager
	kubeproxy  *kubeproxy.Kubeproxy
}

func NewKubelet() *Kubelet {
	kubelet := &Kubelet{}
	kubelet.podManager = podManager.NewPodManager()
	kubelet.kubeproxy = kubeproxy.NewKubeproxy()
	return kubelet
}

func (k *Kubelet) AddPodFromConfig(config module.Config) error {
	return k.podManager.AddPodFromConfig(config)
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
	HandlePodAdditions(pods []*pod.Pod)
	HandlePodUpdates(pods []*pod.Pod)
	HandlePodRemoves(pods []*pod.Pod)
	HandlePodReconcile(pods []*pod.Pod)
	HandlePodSyncs(pods []*pod.Pod)
	HandlePodCleanups() error
}

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
