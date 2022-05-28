package object

type VirtualService struct {
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec       VirtualServiceSpec `json:"spec" yaml:"spec"`
}

type VirtualServiceSpec struct {
	Host string       `json:"hosts" yaml:"hosts"`
	Http []*HTTPRoute `json:"route" yaml:"route"`
}

type HTTPRoute struct {
	Name  string                  `json:"name" yaml:"name"`
	Match []*HTTPMatchRequest     `json:"match" yaml:"match"`
	Route []*HTTPRouteDestination `json:"destination" yaml:"destination"`
}

type HTTPMatchRequest struct {
	Uri       *string `json:"uri" yaml:"uri"`
	PrefixUri *string `json:"prefix" yaml:"prefix"`
	RegexUri  *string `json:"regex" yaml:"regex"`
}

type Destination struct {
	Host   string `json:"host" yaml:"host"`
	Subset string `json:"subset" yaml:"subset"`
}

type HTTPRouteDestination struct {
	Destination *Destination
	Weight      int32
}
