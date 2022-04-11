package main

import (
	"fmt"
	"minik8s/cmd/kubelet/app/config"
	"minik8s/cmd/kubelet/app/pod"
	"minik8s/cmd/kubelet/app/types"
)

type Kubelet struct {
	podConfig *config.PodConfig
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
