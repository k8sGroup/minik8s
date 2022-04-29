package podManager

import (
	"errors"
	"github.com/pquerna/ffjson/ffjson"
	"minik8s/cmd/kubelet/app/dockerClient"
	"minik8s/cmd/kubelet/app/message"
	"minik8s/cmd/kubelet/app/module"
	"minik8s/cmd/kubelet/app/pod"
	"minik8s/pkg/klog"
)

//存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	uid2pod   map[string]*pod.Pod //uid-pod 的映射
	name2uuid map[string]string   //name-uuid的映射
	//rwLock       sync.RWMutex
}

var instance *PodManager

func NewPodManager() *PodManager {
	newManager := &PodManager{}
	//var rwLock sync.RWMutex
	//newManager.rwLock = rwLock
	newManager.uid2pod = make(map[string]*pod.Pod)
	newManager.name2uuid = make(map[string]string)
	return newManager
}

func GetPodManager() *PodManager {
	if instance == nil {
		instance = new(PodManager)
		return instance
	} else {
		return instance
	}
}

func (p *PodManager) GetPodInfo(podName string) ([]byte, error) {
	//p.rwLock.RLock()
	//defer p.rwLock.RUnlock()
	uid, ok := p.name2uuid[podName]
	if !ok {
		err := errors.New(podName + "对应的pod不存在")
		return nil, err
	}
	pod, _ := p.uid2pod[uid]
	res := pod.GetPodSnapShoot()
	return ffjson.Marshal(res)
}
func (p *PodManager) AddPodFromConfig(config module.Config) {
	//p.rwLock.Lock()
	//defer p.rwLock.Unlock()
	//form containers
	newPod := pod.NewPodfromConfig(&config)
	newPodShoot := newPod.GetPodSnapShoot()
	p.uid2pod[newPodShoot.Uid] = newPod
	p.name2uuid[newPodShoot.Name] = newPodShoot.Uid
	//newPod.LabelMap["app"] = config.MetaData.Labels.App
	commandWithConfig := &message.CommandWithConfig{}
	commandWithConfig.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	commandWithConfig.Group = config.Spec.Containers
	//把config中的container里的volumeMounts MountPath 换成实际路径
	for _, value := range commandWithConfig.Group {
		if value.VolumeMounts != nil {
			for index, it := range value.VolumeMounts {
				path, ok := newPodShoot.TmpDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPodShoot.HostDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPodShoot.HostFileMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				klog.Errorf("container Mount path didn't exist")
			}
		}
	}
	podCommand := message.PodCommand{
		ContainerCommand: &(commandWithConfig.Command),
		PodUid:           newPodShoot.Uid,
		PodCommandType:   message.ADD_POD,
	}
	//塞进对应pod的commandChan
	newPod.ReceivePodCommand(podCommand)
	return
}

func (p *PodManager) PullImages(images []string) error {
	commandWithImages := &message.CommandWithImages{}
	commandWithImages.CommandType = message.COMMAND_PULL_IMAGES
	commandWithImages.Images = images
	response := dockerClient.HandleCommand(&(commandWithImages.Command))
	return response.Err
}
