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

//从子网ip到浮动Ip的映射
var GlobalIpMap = map[string]string{
	"192.168.1.4":  "10.119.11.159",
	"192.168.1.6":  "10.119.11.151",
	"192.168.1.10": "10.119.11.144",
	"192.168.1.7":  "10.119.11.164",
}

func FormVxLanPort() string {
	lock.Lock()
	defer lock.Unlock()
	name := VXLAN_PORT_NAME + fmt.Sprintf("%d", count)
	count++
	return name
}
