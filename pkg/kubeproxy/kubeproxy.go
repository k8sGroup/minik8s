package kubeproxy

import (
	"errors"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/kubeproxy/iptablesManager"
	"sync"
)

type Kubeproxy struct {
	portMappings []iptablesManager.PortMapping
	rwLock       sync.RWMutex
	
	//存一份snapshoop
	kubeproxySnapShoot KubeproxySnapShoot
}
type KubeproxySnapShoot struct {
	PortMappings []iptablesManager.PortMapping
}

func NewKubeproxy() *Kubeproxy {
	newKubeproxy := &Kubeproxy{}
	var rwLock sync.RWMutex
	newKubeproxy.rwLock = rwLock
	newKubeproxy.kubeproxySnapShoot = KubeproxySnapShoot{
		PortMappings: newKubeproxy.portMappings,
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
