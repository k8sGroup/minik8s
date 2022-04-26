package pod

import (
	"github.com/satori/go.uuid"
	"minik8s/cmd/kubelet/app/message"
	"minik8s/cmd/kubelet/app/module"
	"minik8s/cmd/kubelet/app/podWorker"
	"minik8s/pkg/klog"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

const emptyDir = "emptyDir"
const hostPath = "hostPath"

//pod的状态
const POD_PENDING_STATUS = "Pending"
const POD_FAILED_STATUS = "Failed"
const POD_RUNNING_STATUS = "Running"
const POD_EXITED_STATUS = "Exited"
const POD_DELETED_STATUS = "deleted"

//container 的状态
const CONTAINER_EXITED_STATUS = "exited"
const CONTAINER_RUNNING_STATUS = "running"
const CONTAINER_CREATED_STATUS = "created"

//pod探针间隔,为了防止探针command拥堵，要等上一次的response
const PROBE_INTERVAL = 60 //探针间隔，单位为秒
type Pod struct {
	name string
	//LabelMap     map[string]string
	uid string
	//create time
	ctime       string
	containers  []module.ContainerMeta
	tmpDirMap   map[string]string
	hostDirMap  map[string]string
	hostFileMap map[string]string
	status      string
	//如果有错的错误信息
	err error
	//读写锁，更小粒度
	rwLock       sync.RWMutex
	commandChan  chan message.PodCommand
	responseChan chan message.PodResponse
	podWorker    *podWorker.PodWorker
	//探针相关
	timer        *time.Ticker
	canProbeWork bool
	stopChan     chan bool
	//存一份snapshot
	podSnapShoot PodSnapShoot
}

//获取pod的信息快照
type PodSnapShoot struct {
	Name        string
	Uid         string
	Ctime       string
	Containers  []module.ContainerMeta
	TmpDirMap   map[string]string
	HostDirMap  map[string]string
	HostFileMap map[string]string
	Status      string
	Err         string
}

//------初始化相关函数--------//
func NewPodfromConfig(config *module.Config) *Pod {
	newPod := &Pod{}
	newPod.name = config.MetaData.Name
	newPod.uid = uuid.NewV4().String()
	newPod.ctime = time.Now().String()
	newPod.canProbeWork = true
	var rwLock sync.RWMutex
	newPod.rwLock = rwLock
	newPod.commandChan = make(chan message.PodCommand, 100)
	newPod.responseChan = make(chan message.PodResponse, 100)
	newPod.podWorker = &podWorker.PodWorker{}
	//创建pod里的containers同时把config里的originName替换为realName
	for index, value := range config.Spec.Containers {
		realName := newPod.name + "_" + value.Name
		newPod.containers = append(newPod.containers, module.ContainerMeta{
			OriginName: value.Name,
			RealName:   realName,
		})
		config.Spec.Containers[index].Name = realName
	}
	newPod.AddVolumes(config.Spec.Volumes)
	newPod.status = POD_PENDING_STATUS //此时还未部署,设置状态为Pending
	//生成snapShoot
	errMsg := ""
	if newPod.err != nil {
		errMsg = newPod.err.Error()
	}
	newPod.podSnapShoot = PodSnapShoot{
		Name:        newPod.name,
		Uid:         newPod.uid,
		Ctime:       newPod.ctime,
		Containers:  newPod.containers,
		TmpDirMap:   newPod.tmpDirMap,
		HostFileMap: newPod.hostFileMap,
		HostDirMap:  newPod.hostDirMap,
		Status:      newPod.status,
		Err:         errMsg,
	}
	//启动pod
	newPod.StartPod()
	return newPod
}
func (p *Pod) StartPod() {
	go p.podWorker.SyncLoop(p.commandChan, p.responseChan)
	go p.listeningResponse()
	p.StartProbe()
}

func (p *Pod) listeningResponse() {
	//删除pod后释放资源
	defer p.releaseResource()
	for {
		select {
		case response, ok := <-p.responseChan:
			if !ok {
				return
			}
			switch response.PodResponseType {
			case message.ADD_POD:
				//判断是否成功
				p.rwLock.Lock()
				responseWithContainIds := (*message.ResponseWithContainIds)(unsafe.Pointer(response.ContainerResponse))
				if responseWithContainIds.Err != nil {
					//出错了
					p.SetStatusAndErr(POD_FAILED_STATUS, responseWithContainIds.Err)
					klog.Errorf(responseWithContainIds.Err.Error())
				} else {
					//设置containersId
					p.SetContainers(responseWithContainIds.Containers, POD_RUNNING_STATUS)
				}
				p.rwLock.Unlock()
			case message.PROBE_POD:
				p.rwLock.Lock()
				if p.status == POD_DELETED_STATUS {
					p.canProbeWork = true
					p.rwLock.Unlock()
				} else {
					responseWithProbeInfos := (*message.ResponseWithProbeInfos)(unsafe.Pointer(response.ContainerResponse))
					if responseWithProbeInfos.Err != nil {
						p.err = responseWithProbeInfos.Err
						p.status = POD_FAILED_STATUS
					} else {
						p.status = POD_RUNNING_STATUS
						for _, value := range responseWithProbeInfos.ProbeInfos {
							if value == CONTAINER_CREATED_STATUS {
								p.status = POD_PENDING_STATUS
								break
							}
							if value == CONTAINER_EXITED_STATUS {
								p.status = POD_EXITED_STATUS
								break
							}
						}
					}
					p.canProbeWork = true
					p.rwLock.Unlock()
				}

			case message.DELETE_POD:
				return
			}
		}
	}
}
func (p *Pod) ReceivePodCommand(podCommand message.PodCommand) {
	p.commandChan <- podCommand
}
func (p *Pod) AddVolumes(volumes []module.Volume) error {
	p.tmpDirMap = make(map[string]string)
	p.hostDirMap = make(map[string]string)
	p.hostFileMap = make(map[string]string)
	for _, value := range volumes {
		if value.Type == emptyDir {
			//临时目录，随机生成
			u := uuid.NewV4()
			path := GetCurrentAbPathByCaller() + "/tmp/" + u.String()
			os.MkdirAll(path, os.ModePerm)
			p.tmpDirMap[value.Name] = path
		} else if value.Type == hostPath {
			//指定了实际目录
			_, err := os.Stat(value.Path)
			if err != nil {
				os.MkdirAll(value.Path, os.ModePerm)
			}
			p.hostDirMap[value.Name] = value.Path
		} else {
			//文件映射
			_, err := os.Stat(value.Path)
			if err != nil {
				return err
			}
			p.hostFileMap[value.Name] = value.Path
		}
	}
	return nil
}

//获取当前文件的路径，
func GetCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

//----------------初始化相关函数结束------------------------//

//----------------辅助函数--------------------------------//
func (p *Pod) SetStatusAndErr(status string, err error) {
	p.status = status
	p.err = err
}
func (p *Pod) SetContainers(containers []module.ContainerMeta, status string) {
	for _, value := range containers {
		for index, it := range p.containers {
			if it.RealName == value.RealName {
				p.containers[index].ContainerId = value.ContainerId
			}
		}
	}
	p.status = status
}

//-------------------------------------------------------//

//-----------------读取pod信息，需要读锁------------------------//
func (p *Pod) GetPodSnapShoot() PodSnapShoot {
	//p.rwLock.TryRLock()
	//defer p.rwLock.RUnlock()
	//只有获取到锁的情况下才会更新 podSnapshoot并返回新值，否者返回旧缓存
	if p.rwLock.TryRLock() {
		errMsg := ""
		if p.err != nil {
			errMsg = p.err.Error()
		}
		p.podSnapShoot = PodSnapShoot{
			Name:        p.name,
			Uid:         p.uid,
			Ctime:       p.ctime,
			Containers:  p.containers,
			TmpDirMap:   p.tmpDirMap,
			HostFileMap: p.hostFileMap,
			HostDirMap:  p.hostDirMap,
			Status:      p.status,
			Err:         errMsg,
		}
		p.rwLock.RUnlock()
		return p.podSnapShoot
	} else {
		return p.podSnapShoot
	}

}

//-----------------------------------------------------------//

//--------------写pod信息，需要写锁------------------------------//
func (p *Pod) StartProbe() {
	p.timer = time.NewTicker(PROBE_INTERVAL * time.Second)
	p.stopChan = make(chan bool)
	go func(p *Pod) {
		defer p.timer.Stop()
		for {
			select {
			case <-p.timer.C:
				p.rwLock.Lock()
				//这几种情况下不进行检查
				if p.canProbeWork && p.status != POD_PENDING_STATUS && p.status != POD_FAILED_STATUS && p.status != POD_DELETED_STATUS {
					command := &message.CommandWithContainerIds{}
					command.CommandType = message.COMMAND_PROBE_CONTAINER
					var group []string
					for _, value := range p.containers {
						group = append(group, value.ContainerId)
					}
					command.ContainerIds = group
					podCommand := message.PodCommand{
						PodUid:           p.uid,
						PodCommandType:   message.PROBE_POD,
						ContainerCommand: &(command.Command),
					}
					p.commandChan <- podCommand
					p.canProbeWork = false
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

func (p *Pod) DeletePod() {
	p.rwLock.Lock()
	p.status = POD_DELETED_STATUS
	command := &message.CommandWithContainerIds{}
	command.CommandType = message.COMMAND_DELETE_CONTAINER
	var group []string
	for _, value := range p.containers {
		group = append(group, value.ContainerId)
	}
	command.ContainerIds = group
	podCommand := message.PodCommand{
		PodUid:           p.uid,
		PodCommandType:   message.DELETE_POD,
		ContainerCommand: &(command.Command),
	}
	p.commandChan <- podCommand
	p.rwLock.Unlock()
}

//释放所有资源
func (p *Pod) releaseResource() {
	//拿下锁防止寄了
	p.rwLock.Lock()
	p.canProbeWork = false
	p.stopChan <- true
	close(p.commandChan)
	close(p.responseChan)
	p.rwLock.Unlock()
}

//-----------------------------------------------------------//
