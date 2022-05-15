package netconfig

import (
	"fmt"
	"sync"
)

const ETH_NAME = "ens3"
const OVS_BRIDGE_NAME = "br0"
const VXLAN_PORT_NAME = "vxlan"
const DOCKER_NETCARD = "docker0"
const DOCKER_IPTABLE_CHAIN = "DOCKER"

var count = 0
var lock sync.Mutex

func FormVxLanPort() string {
	lock.Lock()
	defer lock.Unlock()
	name := VXLAN_PORT_NAME + fmt.Sprintf("%d", count)
	count++
	return name
}
