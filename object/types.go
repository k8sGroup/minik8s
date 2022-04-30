package object

const (
	PodPending   string = "Pending"
	PodRunning   string = "Running"
	PodSucceeded string = "Succeeded"
	PodFailed    string = "Failed"
	PodUnknown   string = "Unknown"

	// SUCCESS http status code
	SUCCESS int = 200
	FAILED  int = 400
)

type ObjectMeta struct {
	Name string
}

/*******************ReplicaSet*************************/

// ReplicaSet ensures that a specified number of pod replicas are running at any given time.
type ReplicaSet struct {
	ObjectMeta
	Spec   ReplicaSetSpec
	Status ReplicaSetStatus
}

type ReplicaSetSpec struct {
	Replicas *int32
	Template PodTemplateSpec
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	Replicas int32
}

type LabelSelector struct {
}

/*******************Pod*************************/

type Pod struct {
	ObjectMeta
	Spec   PodSpec
	Status PodStatus
}

type PodSpec struct {
	NodeName string
}

type PodStatus struct {
	Phase string
	// IP address when the pod is assigned
	HostIP string
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string
}

type PodTemplateSpec struct {
	Spec PodSpec
}

// PodList is a list of Pods.
type PodList struct {
}

type ListOptions struct {
}

/*******************Node*************************/

type Node struct {
	ObjectMeta
	Spec   NodeSpec
	Status NodeStatus
}

type NodeList struct {
	Items []Node
}

type NodeSpec struct {
}

type NodeStatus struct {
}
