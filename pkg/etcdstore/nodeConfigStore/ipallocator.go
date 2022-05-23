package nodeConfigStore

import (
	"fmt"
	"sync"
)

//这个是大网段，确保为node分配的网段全在这里边
const BASIC_IP_AND_MASK = "172.16.0.0/16"
const NODE_NAME_PREFIX = "node"
const MasterNodeName = "node0"

var count = 1
var lock sync.Mutex

//第一个分配的就是docker0的初始地址，也就是说master不需要重启docker0
func GetNodeName() string {
	lock.Lock()
	defer lock.Unlock()
	nodeName := NODE_NAME_PREFIX + fmt.Sprintf("%d", count)
	count++
	return nodeName
}
