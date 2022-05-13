package netconfig

import (
	"fmt"
	"sync"
)

const ETH_NAME = "ens3"
const OVS_BRIDGE_NAME = "br0"
const GRE_PORT_NAME = "gre"
const DOCKER_NETCARD = "docker0"
const DOCKER_IPTABLE_CHAIN = "DOCKER"

var count = 0
var lock sync.Mutex

func FormGrePort() string {
	lock.Lock()
	defer lock.Unlock()
	name := GRE_PORT_NAME + fmt.Sprintf("%d", count)
	count++
	return name
}
