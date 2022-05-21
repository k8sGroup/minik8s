package object

type VersionedDeployment struct {
	Version    int64
	Deployment Deployment
}

func SelectNewerDeployment(d1 VersionedDeployment, d2 VersionedDeployment) VersionedDeployment {
	if d1.Version > d2.Version {
		return d1
	} else {
		return d2
	}
}

type VersionedReplicaset struct {
	Version    int64
	Replicaset ReplicaSet
}

func SelectNewerReplicaset(rs1 VersionedReplicaset, rs2 VersionedReplicaset) VersionedReplicaset {
	if rs1.Version > rs2.Version {
		return rs1
	} else {
		return rs2
	}
}

type VersionedAutoscaler struct {
	Version    int64
	Autoscaler Autoscaler
}

func SelectNewerAutoscaler(rs1 VersionedAutoscaler, rs2 VersionedAutoscaler) VersionedAutoscaler {
	if rs1.Version > rs2.Version {
		return rs1
	} else {
		return rs2
	}
}

type VersionedGPUJob struct {
	Version int64
	Job     GPUJob
}

func SelectNewerGPUJob(vj1 VersionedGPUJob, vj2 VersionedGPUJob) VersionedGPUJob {
	if vj1.Version > vj2.Version {
		return vj1
	} else {
		return vj2
	}
}

type VersionedJobStatus struct {
	Version   int64
	JobStatus JobStatus
}

func SelectNewerJobStatus(vj1 VersionedJobStatus, vj2 VersionedJobStatus) VersionedJobStatus {
	if vj1.Version > vj2.Version {
		return vj1
	} else {
		return vj2
	}
}
