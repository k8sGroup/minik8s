package nodeConfigStore

import (
	"errors"
	"fmt"
	"minik8s/object"
	"sync"
	"time"
)

//内存中的网络信息用于controller的直接使用, ectd中的配置信息用于其他节点的操作
//physical Ip应该是对用户侧封闭的，用户更多应该使用node name，更加友好
type NetConfigStore struct {
	//ipPairs []IpPair
	Name2Node map[string]*object.Node
	//大二层统一网段
	BasicIpAndPair string
}

var instance *NetConfigStore
var rwLock sync.RWMutex

func getNetConfigStore() *NetConfigStore {
	if instance == nil {
		instance = newNetConfigStore()
		return instance
	} else {
		return instance
	}
}

func newNetConfigStore() *NetConfigStore {
	res := &NetConfigStore{}
	res.Name2Node = make(map[string]*object.Node)
	res.BasicIpAndPair = BASIC_IP_AND_MASK
	return res
}
func GetNodes() []*object.Node {
	rwLock.RLock()
	defer rwLock.RUnlock()
	var res []*object.Node
	netConfigStore := getNetConfigStore()
	for _, val := range netConfigStore.Name2Node {
		res = append(res, val)
	}
	return res
}
func AddNewNode(physicalIp string) (*object.Node, error) {
	rwLock.Lock()
	defer rwLock.Unlock()
	//先检查是否已经存在
	netConfigStore := getNetConfigStore()
	flag := netConfigStore.checkIfExist(physicalIp)
	if flag {
		return nil, errors.New(physicalIp + " already exist")
	}
	name, ipAndMask := GetNodeNameWithIpAndMask()
	newNode := &object.Node{
		Spec: object.NodeSpec{
			PhysicalIp:    physicalIp,
			NodeIpAndMask: ipAndMask,
		},
		MetaData: object.ObjectMeta{
			Name:  name,
			Ctime: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	netConfigStore.Name2Node[name] = newNode
	fmt.Println("new node added")
	fmt.Println(*netConfigStore)
	return newNode, nil
}

func DeleteNode(physicalIp string) (*object.Node, error) {
	rwLock.Lock()
	defer rwLock.Unlock()
	netConfigStore := getNetConfigStore()
	flag := netConfigStore.checkIfExist(physicalIp)
	if !flag {
		return nil, errors.New(physicalIp + " not exist")
	}
	var del *object.Node
	for k, v := range netConfigStore.Name2Node {
		if v.Spec.PhysicalIp == physicalIp {
			del = v
			delete(netConfigStore.Name2Node, k)
			return del, nil
		}
	}
	return del, nil

}

func (netConfigStore *NetConfigStore) checkIfExist(physicalIp string) bool {
	for _, v := range netConfigStore.Name2Node {
		if v.Spec.PhysicalIp == physicalIp {
			return true
		}
	}
	return false
}
