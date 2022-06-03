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
	Name      string            `json:"name" yaml:"name"`
	MatchType int               `json:"match_type" yaml:"match_type"` // 0 regex 1 weight
	PDest     []*PodDestination `json:"pdest" yaml:"pdest"`
}

type PodDestination struct {
	PodIP  string `json:"podIP" yaml:"podIP"`
	Weight int32  `json:"weight" yaml:"weight"` // match by weight
	Uri    string `json:"uri" yaml:"uri"`       // match by http request uri
}

type SidecarInject struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Inbound    int  `json:"inbound" yaml:"inbound"`
	Outbound   int  `json:"outbound" yaml:"outbound"`
	SysUid     int  `json:"sysUid" yaml:"sysUid"`
	Status     bool `json:"status" yaml:"status"`
}
