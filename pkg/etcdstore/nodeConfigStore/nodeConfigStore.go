package nodeConfigStore

import (
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
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

//主要任务是分配一个
func AddNewNode(node *object.Node) (*object.Node, error) {
	rwLock.Lock()
	defer rwLock.Unlock()
	//先检查是否已经存在
	netConfigStore := getNetConfigStore()
	flag := netConfigStore.checkIfExist(node.Spec.DynamicIp)
	if flag {
		return nil, errors.New("[nodeConfigStore]" + node.Spec.DynamicIp + " already exist")
	}
	name := GetNodeName()
	node.MetaData.Name = name
	node.MetaData.Ctime = time.Now().Format("2006-01-02 15:04:05")
	node.MetaData.UID = uuid.NewV4().String()
	netConfigStore.Name2Node[name] = node
	fmt.Println("new node added")
	return node, nil
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
		if v.Spec.DynamicIp == physicalIp {
			del = v
			delete(netConfigStore.Name2Node, k)
			return del, nil
		}
	}
	return del, nil

}

func (netConfigStore *NetConfigStore) checkIfExist(physicalIp string) bool {
	for _, v := range netConfigStore.Name2Node {
		if v.Spec.DynamicIp == physicalIp {
			return true
		}
	}
	return false
}
