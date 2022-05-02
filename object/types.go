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
	Volumes    []Volume
	Containers []Container
	NodeName   string `json:"nodeName" yaml:"nodeName"`
}

type PodStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// IP address when the pod is assigned
	HostIP string `json:"hostIP" yaml:"hostIP"`
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string `json:"podIP" yaml:"podIP"`
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
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type Container struct {
	Name         string        `yaml:"name"`
	Image        string        `yaml:"image"`
	Command      []string      `yaml:"command"`
	Args         []string      `yaml:"args"`
	VolumeMounts []VolumeMount `yaml:"volumeMounts"`
	Limits       Limit         `yaml:"limits"`
	Ports        []Port        `yaml:"ports"`
	Env          []EnvEntry    `yaml:"env"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}

type ContainerMeta struct {
	OriginName  string
	RealName    string
	ContainerId string
}

type Limit struct {
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type Port struct {
	ContainerPort string `yaml:"containerPort"`
}

type EnvEntry struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
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
