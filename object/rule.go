package object

type VirtualService struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       VirtualServiceSpec `json:"spec" yaml:"spec"`
}

type VirtualServiceSpec struct {
	Host  string `json:"hosts" yaml:"hosts"` // service name
	Route Route  `json:"route" yaml:"route"`
}

type Route struct {
	Name  string            `json:"name" yaml:"name"`
	PDest []*PodDestination `json:"pdest" yaml:"pdest"`
}

type PodDestination struct {
	PodIP  string `json:"podIP" yaml:"podIP"` // host can use regex
	Weight int32  `json:"weight" yaml:"weight"`
}
