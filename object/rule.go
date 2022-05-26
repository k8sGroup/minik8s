package object

type VirtualService struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       VirtualServiceSpec `json:"spec" yaml:"spec"`
}

type VirtualServiceSpec struct {
	Hosts []string `json:"hosts" yaml:"hosts"`
	Http  []*HTTPRoute
}

type HTTPRoute struct {
	Route []*HTTPRouteDestination
}

type Destination struct {
	Host   string        `json:"host" yaml:"host"`
	Subset string        `json:"subset" yaml:"subset"`
	Port   *PortSelector `json:"port" yaml:"port"`
}

type HTTPRouteDestination struct {
	Destination *Destination
	Weight      int32
}

type PortSelector struct {
	Number uint32
}
