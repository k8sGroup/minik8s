package object

const (
	PodPending   string = "Pending"
	Running      string = "Running"
	PodSucceeded string = "Succeeded"
	Failed       string = "Failed"
	PodUnknown   string = "Unknown"
	Delete       string = "Delete"
	PodExit      string = "Exited"

	// SUCCESS http status code
	SUCCESS int = 200
	FAILED  int = 400

	// kind
	PodKind        string = "Pod"
	ReplicaSetKind string = "RS"

	// Node
	NodeShardFilePath string = "/home/sharedData"
)

type ObjectMeta struct {
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels" yaml:"labels"`
	UID    string            `json:"uid" yaml:"uid"`

	OwnerReferences []OwnerReference `json:"ownerReferences" yaml:"ownerReferences"`
	Ctime           string
}

// OwnerReference ownership for objects, e.g. replicaset and pods
type OwnerReference struct {
	Kind string `json:"kind" yaml:"kind"`
	Name string `json:"name" yaml:"name"`
	UID  string `json:"uid" yaml:"uid"`
	// If true, this reference points to the managing controller.
	Controller bool `json:"controller" yaml:"controller"`
}

/*******************ReplicaSet*************************/

// ReplicaSet ensures that a specified number of pod replicas are running at any given time.
type ReplicaSet struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       ReplicaSetSpec   `json:"spec" yaml:"spec"`
	Status     ReplicaSetStatus `json:"status" yaml:"status"`
}

type ReplicaSetSpec struct {
	Replicas int32       `json:"replicas" yaml:"replicas"`
	Template PodTemplate `json:"template" yaml:"template"`
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	Replicas int32 `json:"replicas" yaml:"replicas"`
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
	Volumes    []Volume    `json:"volumes" yaml:"volumes"`
	Containers []Container `json:"containers" yaml:"containers"`
	NodeName   string      `json:"nodeName" yaml:"nodeName"`
}

type PodStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// IP address allocated to the pod. Routable at least within the cluster
	PodIP string `json:"podIP" yaml:"podIP"`
	//error message
	Err string `json:"err" yaml:"err"`
}

type PodTemplate struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       PodSpec `json:"spec" yaml:"spec"`
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
	Name         string        `json:"name" yaml:"name"`
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
	//??????????????? tcp, udp, all
	//?????????tcp, all??????????????????
	Protocol string `json:"protocol" yaml:"protocol"`
}

type EnvEntry struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

/*******************Node*************************/

type Node struct {
	MetaData ObjectMeta `json:"metadata" yaml:"metadata"`
	MasterIp string     `json:"masterIp" yaml:"masterIp"`
	Spec     NodeSpec   `json:"spec" yaml:"spec"`
	Status   NodeStatus `json:"status" yaml:"status"`
}

type NodeList struct {
	Items []Node `json:"items" yaml:"items"`
}

type NodeSpec struct {
	//??????ip??????
	DynamicIp string `json:"physicalIp" yaml:"physicalIp""`
	//?????????????????????pod??????
	NodeIpAndMask string `json:"nodeIpAndMask" yaml:"nodeIpAndMask"`
}

type NodeStatus struct {
}

/****************Service****************************/
const (
	ClusterIp string = "ClusterIp"
	NodePort  string = "NodePort"
)

type Service struct {
	MetaData ObjectMeta    `json:"metadata" yaml:"metadata"`
	Spec     ServiceSpec   `json:"spec" yaml:"spec"`
	Status   ServiceStatus `json:"status" yaml:"status"`
}
type ServiceSpec struct {
	//service ???????????? ???ClusterIp??? NodePort??????,?????????ClusterIp,???????????????ClusterIp
	Type string `json:"type" yaml:"type"`
	//????????????Ip????????? ?????????????????????????????????????????????
	ClusterIp string `json:"clusterIp" yaml:"clusterIp"`
	//service???????????????????????????
	Ports []ServicePort `json:"ports" yaml:"ports"`
	//selector
	Selector map[string]string `json:"selector" yaml:"selector"`
	//?????????podsIp
	PodNameAndIps []PodNameAndIp `json:"podNameAndIps"`
}
type ServicePort struct {
	//???????????????
	Name string `json:"name" yaml:"name"`
	//????????????, ??????TCP???UDP, ??????TCP
	Protocol string `json:"protocol" yaml:"protocol"`
	//????????????????????????
	Port string `json:"port" yaml:"port"`
	//?????????????????????Pod????????????
	TargetPort string `json:"targetPort" yaml:"targetPort"`
	//???service?????????NodePort??????????????????????????????????????????
	NodePort string `json:"nodePort" yaml:"nodePort"`
}
type PodNameAndIp struct {
	Name string `json:"name"`
	Ip   string `json:"ip"`
}
type ServiceStatus struct {
	//runtime
	Err string `json:"err" yaml:"err"`
	//pod name??? podIp:port?????????
	Pods2IpAndPort map[string]string `json:"pods2IpAndPort" yaml:"pods2IpAndPort"`
	Phase          string            `json:"phase" yaml:"phase"`
}

/***************************DNS?????????*******************************/
type DnsAndTrans struct {
	MetaData ObjectMeta        `json:"metadata" yaml:"metadata"`
	Spec     DnsAndTransSpec   `json:"spec" yaml:"spec"`
	Status   DnsAndTransStatus `json:"status" yaml:"status"`
}
type DnsAndTransSpec struct {
	Host string `json:"host" yaml:"host"`
	//???????????????gateway??????ip?????????
	GateWayIp string      `json:"gateWayIp" yaml:"gateWayIp"`
	Paths     []TransPath `json:"paths" yaml:"paths"`
}
type TransPath struct {
	Name    string `json:"name" yaml:"name"`
	Service string `json:"service" yaml:"service"`
	Ip      string `json:"ip" yaml:"ip"` //?????????????????????????????????Ip, handler??????
	Port    string `json:"port"  yaml:"port"`
}

const (
	FileCreated    string = "FileCreated"
	ServiceCreated string = "ServiceCreated"
)

type DnsAndTransStatus struct {
	//????????? "", "FileCreated", "ServiceCreated", "deleted"
	Phase string `json:"phase" yaml:"phase"`
}
