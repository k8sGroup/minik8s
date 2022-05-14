package podManager

import (
	"errors"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet/dockerClient"
	"minik8s/pkg/kubelet/message"
	"minik8s/pkg/kubelet/pod"
	"sync"
	"time"
)

//定时更新间隔
const PODMANAGER_TIME_INTERVAL = 20

//存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	uid2pod   map[string]*pod.Pod //uid-pod 的映射
	name2uuid map[string]string   //name-uuid的映射
	//hint:通常用户并不会直接调用 pod.GetPodSnapShoot,所有pod信息的获取直接拿etcd的缓存，所以这里只需要定时把pod信息更新上去即可。
	uid2podSnapshoot map[string]pod.PodSnapShoot //uid-podSnapshoot的映射
	//对map的保护
	rwLock sync.RWMutex
	client client.RESTClient
	//定时器相关
	timer    *time.Ticker
	stopChan chan bool
	//podManager中出现的错误
	Err error
}

var instance *PodManager

func NewPodManager(clientConfig client.Config) *PodManager {
	newManager := &PodManager{}
	//var rwLock sync.RWMutex
	//newManager.rwLock = rwLock
	newManager.uid2pod = make(map[string]*pod.Pod)
	newManager.uid2podSnapshoot = make(map[string]pod.PodSnapShoot)
	newManager.name2uuid = make(map[string]string)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newManager.client = restClient
	var rwLock sync.RWMutex
	newManager.rwLock = rwLock
	return newManager
}
func (p *PodManager) StartPodManager() {
	p.startTimer()
}
func (p *PodManager) startTimer() {
	p.timer = time.NewTicker(PODMANAGER_TIME_INTERVAL * time.Second)
	p.stopChan = make(chan bool)
	go func(p *PodManager) {
		defer p.timer.Stop()
		for {
			select {
			case <-p.timer.C:
				p.rwLock.Lock()
				//对于每个pod调用GetSnapShoot更新信息
				for k, v := range p.name2uuid {
					pod := p.uid2pod[v]
					newPodSnapShoot := pod.GetPodSnapShoot()
					if !compareSame(p.uid2podSnapshoot[v], newPodSnapShoot) {
						//有区别产生，需要更新缓存以及etcd
						oldPod, err := p.client.GetPod(k)
						if err != nil {
							p.Err = err
							continue
						}
						//暂时只用更新这两
						oldPod.Status.Phase = newPodSnapShoot.Status
						oldPod.Status.PodIP = newPodSnapShoot.PodNetWork.Ipaddress
						oldPod.Status.Err = newPodSnapShoot.Err
						err = p.client.UpdatePods(oldPod)
						if err != nil {
							p.Err = err
							continue
						}
						p.uid2podSnapshoot[v] = newPodSnapShoot
					}
				}
				p.rwLock.Unlock()
			case stop := <-p.stopChan:
				if stop {
					return
				}
			}
		}
	}(p)
}

//直接通过缓存获取，并且不去更新map， 所有对map的更新都通过定时器进行
func (p *PodManager) GetPodInfo(podName string) ([]byte, error) {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	uid, ok := p.name2uuid[podName]
	if !ok {
		err := errors.New(podName + "对应的pod不存在")
		return nil, err
	}
	res, _ := p.uid2podSnapshoot[uid]
	return ffjson.Marshal(res)
}

func (p *PodManager) GetPodSnapShoot(podName string) (*pod.PodSnapShoot, error) {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	uid, ok := p.name2uuid[podName]
	if !ok {
		err := errors.New(podName + "对应的pod不存在")
		return nil, err
	}
	res, _ := p.uid2podSnapshoot[uid]
	return &res, nil
}

func (p *PodManager) CheckIfPodExist(podName string) bool {
	_, ok := p.name2uuid[podName]
	return ok
}

func (p *PodManager) DeletePod(podName string) error {
	p.rwLock.Lock()
	p.rwLock.Unlock()
	if !p.CheckIfPodExist(podName) {
		//不存在该pod
		return errors.New(podName + "对应的pod不存在")
	}
	uid, _ := p.name2uuid[podName]
	pod_, _ := p.uid2pod[uid]
	fmt.Printf("[DeletePod] Prepare delete pod")
	pod_.DeletePod()
	delete(p.name2uuid, podName)
	delete(p.uid2podSnapshoot, uid)
	delete(p.uid2pod, uid)
	//提交delete pod的请求
	err := p.client.DeletePod(podName)
	if err != nil {
		return err
	}
	return nil
}

func (p *PodManager) AddPod(config *object.Pod) error {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	//首先检查name对应的pod是否存在， 存在的话报错
	if p.CheckIfPodExist(config.ObjectMeta.Name) {
		return errors.New(config.ObjectMeta.Name + "对应的pod已经存在，请先删除原pod")
	}
	newPod := pod.NewPodfromConfig(config)
	newPodShoot := newPod.GetPodSnapShoot()
	p.uid2pod[newPodShoot.Uid] = newPod
	p.uid2podSnapshoot[newPodShoot.Uid] = newPodShoot
	p.name2uuid[newPodShoot.Name] = newPodShoot.Uid
	//把新的pod信息更新到etcd
	//oldPod, err := p.client.GetPod(newPodShoot.Name)
	//if err != nil {
	//	return err
	//}
	config.Status.Phase = newPodShoot.Status
	config.Status.PodIP = newPodShoot.PodNetWork.Ipaddress
	config.ObjectMeta.Ctime = newPodShoot.Ctime
	config.Status.Err = newPodShoot.Err
	//把新的pod信息上传
	err := p.client.UpdatePods(config)
	return err
}

func (p *PodManager) PullImages(images []string) error {
	commandWithImages := &message.CommandWithImages{}
	commandWithImages.CommandType = message.COMMAND_PULL_IMAGES
	commandWithImages.Images = images
	response := dockerClient.HandleCommand(&(commandWithImages.Command))
	return response.Err
}

// CopyUid2pod only copy the pointers in map, check before actual use
func (p *PodManager) CopyUid2pod() map[string]*pod.Pod {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	uuidMap := make(map[string]*pod.Pod)
	for key, val := range p.uid2pod {
		uuidMap[key] = val
	}
	return uuidMap
}

func compareSame(p1 pod.PodSnapShoot, p2 pod.PodSnapShoot) bool {
	//只需要比较会发生变化的
	if p1.Ctime != p2.Ctime {
		return false
	}
	if p1.Uid != p2.Uid {
		return false
	}
	if len(p1.Containers) != len(p2.Containers) {
		return false
	}
	if p1.PodNetWork.GateWay != p2.PodNetWork.GateWay {
		return false
	}
	if p1.PodNetWork.Ipaddress != p2.PodNetWork.Ipaddress {
		return false
	}
	if len(p1.PodNetWork.OpenPortSet) != len(p2.PodNetWork.OpenPortSet) {
		return false
	}
	if p1.Status != p2.Status {
		return false
	}
	if p1.Err != p2.Err {
		return false
	}
	return true
}
