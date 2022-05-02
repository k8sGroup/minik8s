package netConfigStore

import "errors"

//集群ip--分配的网段
type ipPair struct {
	//集群中的物理ip
	ClusterIp string
	//为该物理机分配的pod网段
	NodeIpAndMask string
}

type NetConfigStore struct {
	IpPairs []ipPair
	//大二层统一网段
	BasicIpAndPair string
}

func (netConfigStore *NetConfigStore) AddNewNode(clusterIp string) (ipPair, error) {
	//先检查是否已经存在
	index := netConfigStore.checkIfExist(clusterIp)
	if index != -1 {
		return ipPair{}, errors.New(clusterIp + " already exist")
	}
	newIpPair := ipPair{
		ClusterIp:     clusterIp,
		NodeIpAndMask: GetNodeIpAndMask(),
	}
	netConfigStore.IpPairs = append(netConfigStore.IpPairs, newIpPair)
	return newIpPair, nil
}

func (netConfigStore *NetConfigStore) DeleteNode(clusterIp string) (ipPair, error) {
	index := netConfigStore.checkIfExist(clusterIp)
	if index == -1 {
		return ipPair{}, errors.New(clusterIp + " not exist")
	}
	del := netConfigStore.IpPairs[index]
	netConfigStore.IpPairs = append(netConfigStore.IpPairs[:index], netConfigStore.IpPairs[index+1:]...)
	return del, nil
}

func (netConfigStore *NetConfigStore) checkIfExist(clusterIp string) int {
	for index, value := range netConfigStore.IpPairs {
		if value.ClusterIp == clusterIp {
			return index
		}
	}
	return -1
}
