package kubeproxy

import (
	"errors"
	"minik8s/pkg/etcdstore/netConfigStore"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/kubeproxy/iptablesManager"
	"sync"
)

//--------------常量定义---------------------//
const OP_ADD_ENTRY = 1
const OP_DELETE_ENTRY = 2

//------------------------------------------//
type Kubeproxy struct {
	portMappings []iptablesManager.PortMapping
	rwLock       sync.RWMutex
	//本地存一份netConfigStore,也便于与新的比较得到具体差别
	netConfigStore netConfigStore.NetConfigStore
	//存一份map,由clusterIp映射到管道名
	ipPipeMap map[string]string
	//存一份snapshoop
	kubeproxySnapShoot KubeproxySnapShoot
}

type netConfigStoreEntry struct {
	//操作类型
	Op        int
	ClusterIp string
}
type KubeproxySnapShoot struct {
	PortMappings []iptablesManager.PortMapping
	IpPipeMap    map[string]string
}

func NewKubeproxy() *Kubeproxy {
	newKubeproxy := &Kubeproxy{}
	var rwLock sync.RWMutex
	newKubeproxy.rwLock = rwLock
	newKubeproxy.ipPipeMap = make(map[string]string)
	newKubeproxy.kubeproxySnapShoot = KubeproxySnapShoot{
		PortMappings: newKubeproxy.portMappings,
		IpPipeMap:    newKubeproxy.ipPipeMap,
	}

	return newKubeproxy
}
func (k *Kubeproxy) RemovePortMapping(pod *pod.PodSnapShoot, dockerPort string, hostPort string) error {
	//先检查是否存在该规则
	k.rwLock.Lock()
	defer k.rwLock.Unlock()
	for index, value := range k.portMappings {
		if value.DockerPort == dockerPort && value.HostPort == hostPort && value.DockerIp == pod.PodNetWork.Ipaddress {
			err := iptablesManager.RemoveDockerChainMappingRule(value)
			if err != nil {
				return err
			}
			k.portMappings = append(k.portMappings[:index], k.portMappings[index+1:]...)
			return nil
		}
	}
	return errors.New("想要删除的规则不存在")
}
func (kubeproxy *Kubeproxy) AddPortMapping(pod *pod.PodSnapShoot, dockerPort string, hostPort string) (iptablesManager.PortMapping, error) {
	//要检查pod中是否存在该dockerPort
	kubeproxy.rwLock.Lock()
	defer kubeproxy.rwLock.Unlock()
	flag := false
	for _, value := range pod.PodNetWork.OpenPortSet {
		if value == dockerPort {
			flag = true
			break
		}
	}
	if !flag {
		return iptablesManager.PortMapping{}, errors.New("pod:" + pod.Name + "并未开放端口" + dockerPort)
	}
	//建立PortMapping
	res, err := iptablesManager.AddDockerChainMappingRule(pod.PodNetWork.Ipaddress, dockerPort, hostPort)
	if err != nil {
		return iptablesManager.PortMapping{}, err
	}
	kubeproxy.portMappings = append(kubeproxy.portMappings, res)
	return res, nil
}
func (kubeproxy *Kubeproxy) GetKubeproxySnapShoot() KubeproxySnapShoot {
	if kubeproxy.rwLock.TryRLock() {
		kubeproxy.kubeproxySnapShoot = KubeproxySnapShoot{
			PortMappings: kubeproxy.portMappings,
		}
		kubeproxy.rwLock.RUnlock()
		return kubeproxy.kubeproxySnapShoot
	} else {
		return kubeproxy.kubeproxySnapShoot
	}
}

//TODO 是否要设计成异步的命令处理? 容错处理?
func (kubeproxy *Kubeproxy) compareAndGetDiff(newNetConfigStore netConfigStore.NetConfigStore) []netConfigStoreEntry {
	var res []netConfigStoreEntry
	//先找到delete的entry
	for _, value := range kubeproxy.netConfigStore.IpPairs {
		if isEntryExist(value.ClusterIp, newNetConfigStore) == -1 {
			//not exist, means deleted
			res = append(res, netConfigStoreEntry{
				Op:        OP_DELETE_ENTRY,
				ClusterIp: value.ClusterIp,
			})
		}
	}
	//找到add的entry
	for _, value := range newNetConfigStore.IpPairs {
		if isEntryExist(value.ClusterIp, kubeproxy.netConfigStore) == -1 {
			res = append(res, netConfigStoreEntry{
				Op:        OP_ADD_ENTRY,
				ClusterIp: value.ClusterIp,
			})
		}
	}
	return res
}
func isEntryExist(clusterIp string, store netConfigStore.NetConfigStore) int {
	for index, value := range store.IpPairs {
		if value.ClusterIp == clusterIp {
			return index
		}
	}
	return -1
}
