package kubeproxy

const (
	GeneralServiceChain string = "HONG-SERVICE"
	OutPutChain         string = "OUTPUT"
	PreRoutingChain     string = "PREROUTING"
	NatTable            string = "nat"
	SepChainPrefix      string = "HONG-SEP"
	SvcChainPrefix      string = "HONG-SVC"
	TCP                 string = "tcp"
	UDP                 string = "udp"
)
