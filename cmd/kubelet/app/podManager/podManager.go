package podManager

import (
	uuid "github.com/satori/go.uuid"
	"minik8s/cmd/kubelet/app/dockerClient"
	"minik8s/cmd/kubelet/app/message"
	"minik8s/cmd/kubelet/app/module"
	"minik8s/cmd/kubelet/app/pod"
	"minik8s/pkg/klog"
	"unsafe"
)

//存储所有的pod信息， 当需要获取pod信息时，直接从缓存中取，速度快  需要初始化变量
type PodManager struct {
	uid2pod      map[string]*pod.Pod //uid-pod 的映射
	name2uuid    map[string]string   //name-uuid的映射
	commandChan  chan message.PodCommand
	responseChan chan message.PodResponse
}

var instance *PodManager

func NewPodManager() *PodManager {
	newManager := &PodManager{}
	newManager.uid2pod = make(map[string]*pod.Pod)
	newManager.name2uuid = make(map[string]string)
	newManager.commandChan = make(chan message.PodCommand, 100)
	newManager.responseChan = make(chan message.PodResponse, 100)
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

func (p PodManager) AddPodFromConfig(config module.Config) {
	//form containers
	newPod := &pod.Pod{}
	newPod.Name = config.MetaData.Name
	newPod.Uid = uuid.NewV4().String()
	newPod.AddVolumes(config.Spec.Volumes)
	newPod.Status = pod.POD_PENDING_STATUS //此时还未部署,设置状态为Pending
	p.uid2pod[newPod.Uid] = newPod
	p.name2uuid[newPod.Name] = newPod.Uid
	//newPod.LabelMap["app"] = config.MetaData.Labels.App
	commandWithConfig := &message.CommandWithConfig{}
	commandWithConfig.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	commandWithConfig.Group = config.Spec.Containers
	//把container里的volumeMounts MountPath 换成实际路径
	for _, value := range commandWithConfig.Group {
		if value.VolumeMounts != nil {
			for _, it := range value.VolumeMounts {
				path, ok := newPod.TmpDirMap[it.MountPath]
				if ok {
					it.MountPath = path
					continue
				}
				path, ok = newPod.HostDirMap[it.MountPath]
				if ok {
					it.MountPath = path
					continue
				}
				klog.Errorf("container Mount path didn't exist")
			}
		}
	}
	podCommand := message.PodCommand{
		ContainerCommand: &(commandWithConfig.Command),
		PodUid:           newPod.Uid,
		PodCommandType:   message.ADD_POD,
	}
	//塞进commandChan
	p.commandChan <- podCommand
	return
	//response := dockerClient.HandleCommand(&(commandWithConfig.Command))
	//if response.Err != nil {
	//	return nil, response.Err
	//}
	//responseWithContainersId := (*message.ResponseWithContainIds)(unsafe.Pointer(response))
	//newPod.ContainerIds = responseWithContainersId.ContainersIds
	////p.pods = append(p.pods, newPod)
	//return newPod, nil
}

func (p PodManager) listenResponse() {
	for {
		select {
		case response, _ := <-p.responseChan:
			switch response.PodResponseType {
			case message.ADD_POD:
				//判断是否成功
				responseWithContainIds := (*message.ResponseWithContainIds)(unsafe.Pointer(response.ContainerResponse))
				if responseWithContainIds.Err != nil {
					//出错了
					klog.Errorf(responseWithContainIds.Err.Error())
					p.uid2pod[response.PodUid].Status = pod.POD_FAILED_STATUS
				} else {
					p.uid2pod[response.PodUid].ContainerIds = responseWithContainIds.ContainersIds
					p.uid2pod[response.PodUid].Status = pod.POD_RUNNING_STATUS
				}
			}
		}
	}
}
func (p PodManager) PullImages(images []string) error {
	commandWithImages := &message.CommandWithImages{}
	commandWithImages.CommandType = message.COMMAND_PULL_IMAGES
	commandWithImages.Images = images
	response := dockerClient.HandleCommand(&(commandWithImages.Command))
	return response.Err
}
