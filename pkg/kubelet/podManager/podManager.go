package podManager

import (
	"errors"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet/pod"
	"sync"
)

//定时更新间隔
const PODMANAGER_TIME_INTERVAL = 20

//存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	name2pod map[string]*pod.Pod //name-pod的映射
	//对map的保护
	lock         sync.Mutex
	client       client.RESTClient
	clientConfig client.Config
	Err          error
}

var instance *PodManager

func NewPodManager(clientConfig client.Config) *PodManager {
	newManager := &PodManager{}
	newManager.name2pod = make(map[string]*pod.Pod)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newManager.client = restClient
	newManager.clientConfig = clientConfig
	var lock sync.Mutex
	newManager.lock = lock
	return newManager
}
func (p *PodManager) CheckIfPodExist(podName string) bool {
	_, ok := p.name2pod[podName]
	return ok
}

func (p *PodManager) DeletePod(podName string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.CheckIfPodExist(podName) {
		//不存在该pod
		return errors.New(podName + "对应的pod不存在")
	}
	pod, _ := p.name2pod[podName]
	fmt.Printf("[DeleteRuntimePod] Prepare delete pod")
	pod.DeletePod()
	delete(p.name2pod, podName)
	return nil
}

func (p *PodManager) AddPod(config *object.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	//首先检查name对应的pod是否存在， 存在的话报错
	if p.CheckIfPodExist(config.ObjectMeta.Name) {
		return errors.New(config.ObjectMeta.Name + "对应的pod已经存在，请先删除原pod")
	}
	newPod := pod.NewPodfromConfig(config, p.clientConfig)
	p.name2pod[config.Name] = newPod
	return nil
}

// CopyName2pod only copy the pointers in map, check before actual use
func (p *PodManager) CopyName2pod() map[string]*pod.Pod {
	p.lock.Lock()
	defer p.lock.Unlock()
	uuidMap := make(map[string]*pod.Pod)
	for key, val := range p.name2pod {
		uuidMap[key] = val
	}
	return uuidMap
}
