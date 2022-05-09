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
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels" yaml:"labels"`
	//time to create
	Ctime string
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
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PodSpec   `json:"spec" yaml:"spec"`
	Status     PodStatus `json:"status" yaml:"status"`
}

type PodSpec struct {
	Volumes    []Volume    `json:"volume" yaml:"volume"`
	Containers []Container `json:"containers" yaml:"containers"`
	NodeName   string      `json:"nodeName" yaml:"nodeName"`
}

type PodStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// IP address when the pod is assigned
	HostIP string `json:"hostIP" yaml:"hostIP"`
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string `json:"podIP" yaml:"podIP"`
	//error message
	Err string
}

type PodTemplateSpec struct {
	Spec PodSpec
}

// PodList is a list of Pods.
type PodList struct {
}

type ListOptions struct {
}

type Volume struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
}

type Container struct {
	Name         string        `json:"path" yaml:"name"`
	Image        string        `json:"image" yaml:"image"`
	Command      []string      `json:"command" yaml:"command"`
	Args         []string      `json:"args" yaml:"args"`
	VolumeMounts []VolumeMount `json:"volumeMounts" yaml:"volumeMounts"`
	Limits       Limit         `json:"limits" yaml:"limits"`
	Ports        []Port        `json:"ports" yaml:"ports"`
	Env          []EnvEntry    `json:"env" yaml:"env"`
}

type VolumeMount struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
}

type ContainerMeta struct {
	OriginName  string
	RealName    string
	ContainerId string
}

type Limit struct {
	Cpu    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

type Port struct {
	ContainerPort string `json:"containerPort" yaml:"containerPort"`
}

type EnvEntry struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

/*******************Node*************************/

type Node struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       NodeSpec   `json:"spec" yaml:"spec"`
	Status     NodeStatus `json:"status" yaml:"status"`
}

type NodeList struct {
	Items []Node `json:"items" yaml:"items"`
}

type NodeSpec struct {
}

type NodeStatus struct {
}
