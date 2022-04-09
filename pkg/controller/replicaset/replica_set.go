package replicaset

import (
	"context"
	"minik8s/object"
)

type ReplicaSetController struct {
}

// When a pod is created, enqueue its replica set
func (rsc *ReplicaSetController) addPod(obj interface{}) {

}

// add replicaset to queue
func (rsc *ReplicaSetController) enqueueRS(rs *object.ReplicaSet) {

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
	return nil
}
