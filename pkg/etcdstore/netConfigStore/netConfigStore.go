package netConfigStore

import (
	"errors"
)

//集群ip--分配的网段
type IpPair struct {
	//集群中的物理ip
	ClusterIp string
	//为该物理机分配的pod网段
	NodeIpAndMask string
}

//内存中的网络信息用于controller的直接使用, ectd中的配置信息用于其他节点的操作
//name 是用于用户侧方便显示等等， 实际上唯一标识还是以cluterIp为准(居群中的key)
type NetConfigStore struct {
	//ipPairs []IpPair
	Name2IpPair map[string]IpPair
	//大二层统一网段
	BasicIpAndPair string
}

func NewNetConfigStore() *NetConfigStore {
	res := &NetConfigStore{}
	res.Name2IpPair = make(map[string]IpPair)
	res.BasicIpAndPair = BASIC_IP_AND_MASK
	return res
}

func (netConfigStore *NetConfigStore) AddNewNode(clusterIp string) (IpPair, error) {
	//先检查是否已经存在
	flag := netConfigStore.checkIfExist(clusterIp)
	if flag {
		return IpPair{}, errors.New(clusterIp + " already exist")
	}
	name, ipAndMask := GetNodeNameWithIpAndMask()
	newIpPair := IpPair{
		ClusterIp:     clusterIp,
		NodeIpAndMask: ipAndMask,
	}
	netConfigStore.Name2IpPair[name] = newIpPair
	return newIpPair, nil
}

func (netConfigStore *NetConfigStore) DeleteNode(clusterIp string) (IpPair, error) {
	flag := netConfigStore.checkIfExist(clusterIp)
	if !flag {
		return IpPair{}, errors.New(clusterIp + " not exist")
	}
	var del IpPair
	for k, v := range netConfigStore.Name2IpPair {
		if v.ClusterIp == clusterIp {
			del = v
			delete(netConfigStore.Name2IpPair, k)
			return del, nil
		}
	}
	return del, nil
	//del := netConfigStore.ipPairs[index]
	//netConfigStore.ipPairs = append(netConfigStore.ipPairs[:index], netConfigStore.ipPairs[index+1:]...)

}

func (netConfigStore *NetConfigStore) checkIfExist(clusterIp string) bool {
	for _, v := range netConfigStore.Name2IpPair {
		if v.ClusterIp == clusterIp {
			return true
		}
	}
	return false
}
