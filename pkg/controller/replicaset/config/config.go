package config

// ReplicaSetControllerConfiguration contains elements describing ReplicaSetController.
type ReplicaSetControllerConfiguration struct {
	// concurrentRSSyncs is the number of replica sets that are  allowed to sync
	// concurrently. Larger number = more responsive replica  management, but more
	// CPU (and network) load.
	ConcurrentRSSyncs int32
}
