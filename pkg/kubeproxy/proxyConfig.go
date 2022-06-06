package kubeproxy

const (
	GeneralServiceChain string = "SERVICE"
	OutPutChain         string = "OUTPUT"
	PreRoutingChain     string = "PREROUTING"
	NatTable            string = "nat"
	SepChainPrefix      string = "SEP"
	SvcChainPrefix      string = "SVC"
	TCP                 string = "tcp"
	UDP                 string = "udp"
)
