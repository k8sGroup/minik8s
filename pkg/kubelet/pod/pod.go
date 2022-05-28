package pod

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/satori/go.uuid"
	"minik8s/object"
	"minik8s/pkg/client"
	"minik8s/pkg/kubelet/message"
	"minik8s/pkg/kubelet/podWorker"
	"os"
	"path"
	"runtime"
	"strings"
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
const POD_CREATEED_STATUS = "created"

//container 的状态
const CONTAINER_EXITED_STATUS = "exited"
const CONTAINER_RUNNING_STATUS = "running"
const CONTAINER_CREATED_STATUS = "created"

//pod探针间隔,为了防止探针command拥堵，要等上一次的response
const PROBE_INTERVAL = 60 //探针间隔，单位为秒

type Pod struct {
	configPod   *object.Pod
	containers  []object.ContainerMeta
	tmpDirMap   map[string]string
	hostDirMap  map[string]string
	hostFileMap map[string]string
	//读写锁，更小粒度
	rwLock       sync.RWMutex
	commandChan  chan message.PodCommand
	responseChan chan message.PodResponse
	podWorker    *podWorker.PodWorker
	//探针相关
	timer        *time.Ticker
	canProbeWork bool
	stopChan     chan bool
	client       client.RESTClient
}

type PodNetWork struct {
	OpenPortSet []string //开放端口集合
	GateWay     string   //网关地址 ip4
	Ipaddress   string   //在docker网段中的地址

}

//--------------tool function--------------------------------//
func (p *Pod) GetName() string {
	return p.configPod.Name
}
func (p *Pod) GetLabel() map[string]string{
	return p.configPod.Labels
}
func (p *Pod) GetUid() string {
	return p.configPod.UID
}
func (p *Pod) GetContainers() []object.ContainerMeta {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	deepContainers := p.containers
	return deepContainers
}
//修改了status返回true
func (p *Pod) compareAndSetStatus(status string) bool {
	oldStatus := p.getStatus()
	if oldStatus == status {
		return false
	}
	p.configPod.Status.Phase = status
	return true
}
func (p *Pod) getStatus() string {
	return p.configPod.Status.Phase
}
func (p *Pod) setError(err error) {
	p.configPod.Status.Err = err.Error()
}

func (p *Pod) uploadPod() {
	err := p.client.UpdateRuntimePod(p.configPod)
	if err != nil {
		fmt.Println("[pod] updateRuntimePod error" + err.Error())
	}
}

//--------------------------------------------------------------//

//------初始化相关函数--------//
func NewPodfromConfig(config *object.Pod, clientConfig client.Config) *Pod {
	newPod := &Pod{}
	newPod.configPod = config
	newPod.configPod.Ctime = time.Now().String()
	newPod.canProbeWork = false
	var rwLock sync.RWMutex
	newPod.rwLock = rwLock
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newPod.client = restClient
	newPod.commandChan = make(chan message.PodCommand, 100)
	newPod.responseChan = make(chan message.PodResponse, 100)
	newPod.podWorker = &podWorker.PodWorker{}
	//创建pod里的containers同时把config里的originName替换为realName
	//先填第一个pause容器
	newPod.containers = append(newPod.containers, object.ContainerMeta{
		OriginName: "pause",
		RealName:   "", //先设置为空
	})
	pauseRealName := "pause"
	for index, value := range config.Spec.Containers {
		realName := config.Name + "_" + value.Name
		newPod.containers = append(newPod.containers, object.ContainerMeta{
			OriginName: value.Name,
			RealName:   realName,
		})
		pauseRealName += "_" + realName
		config.Spec.Containers[index].Name = realName
	}
	newPod.containers[0].RealName = pauseRealName
	err := newPod.AddVolumes(config.Spec.Volumes)
	if err != nil {
		newPod.setError(err)
		newPod.compareAndSetStatus(POD_FAILED_STATUS)
	} else {
		newPod.compareAndSetStatus(POD_PENDING_STATUS)
	}
	//启动pod
	newPod.StartPod()
	//生成command
	commandWithConfig := &message.CommandWithConfig{}
	commandWithConfig.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
	commandWithConfig.Group = config.Spec.Containers
	//把config中的container里的volumeMounts MountPath 换成实际路径
	for _, value := range commandWithConfig.Group {
		if value.VolumeMounts != nil {
			for index, it := range value.VolumeMounts {
				path, ok := newPod.tmpDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPod.hostDirMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				path, ok = newPod.hostFileMap[it.Name]
				if ok {
					value.VolumeMounts[index].Name = path
					continue
				}
				fmt.Println("[pod] error:container Mount path didn't exist")
			}
		}
	}
	podCommand := message.PodCommand{
		ContainerCommand: &(commandWithConfig.Command),
		PodCommandType:   message.ADD_POD,
	}
	newPod.commandChan <- podCommand
	//提交pod
	newPod.uploadPod()
	return newPod
}

func (p *Pod) StartPod() {
	go p.podWorker.SyncLoop(p.commandChan, p.responseChan)
	go p.listeningResponse()
	p.canProbeWork = true
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
				fmt.Printf("[pod] receive AddPod responce")
				fmt.Println(*responseWithContainIds)
				fmt.Println(*responseWithContainIds.NetWorkInfos)
				if responseWithContainIds.Err != nil {
					//出错了
					if p.SetStatusAndErr(POD_FAILED_STATUS, responseWithContainIds.Err) {
						p.uploadPod()
					}
					fmt.Println(responseWithContainIds.Err.Error())
				} else {
					//设置containersId
					p.SetContainersAndStatus(responseWithContainIds.Containers, POD_RUNNING_STATUS)
					p.setIpAddress(responseWithContainIds.NetWorkInfos)
					p.uploadPod()
				}
				p.rwLock.Unlock()
			case message.PROBE_POD:
				p.rwLock.Lock()
				if p.getStatus() == POD_DELETED_STATUS {
					p.canProbeWork = false
					p.rwLock.Unlock()
				} else {
					responseWithProbeInfos := (*message.ResponseWithProbeInfos)(unsafe.Pointer(response.ContainerResponse))
					if responseWithProbeInfos.Err != nil {
						p.SetStatusAndErr(POD_FAILED_STATUS, responseWithProbeInfos.Err)
					} else {
						status := POD_RUNNING_STATUS
						for _, value := range responseWithProbeInfos.ProbeInfos {
							if value == CONTAINER_CREATED_STATUS {
								status = POD_CREATEED_STATUS
								break
							}
							if value == CONTAINER_EXITED_STATUS {
								status = POD_EXITED_STATUS
								break
							}
						}
						if p.compareAndSetStatus(status) {
							p.uploadPod()
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

func (p *Pod) AddVolumes(volumes []object.Volume) error {
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
func (p *Pod) SetStatusAndErr(status string, err error) bool {
	p.configPod.Status.Err = err.Error()
	return p.compareAndSetStatus(status)
}
func (p *Pod) SetContainersAndStatus(containers []object.ContainerMeta, status string) bool {
	for _, value := range containers {
		for index, it := range p.containers {
			if it.RealName == value.RealName {
				p.containers[index].ContainerId = value.ContainerId
			}
		}
	}
	return p.compareAndSetStatus(status)
}
func (p *Pod) setIpAddress(settings *types.NetworkSettings) {
	p.configPod.Status.PodIP = settings.IPAddress
}
func filterSingle(input string) string {
	index := strings.Index(input, "/tcp")
	return input[0:index]
}
func filterChars(input []string) []string {
	var result []string
	for _, value := range input {
		result = append(result, filterSingle(value))
	}
	return result
}

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
				if p.canProbeWork && p.getStatus() != POD_PENDING_STATUS && p.getStatus() != POD_FAILED_STATUS && p.getStatus() != POD_DELETED_STATUS && p.getStatus() != POD_EXITED_STATUS {
					command := &message.CommandWithContainerIds{}
					command.CommandType = message.COMMAND_PROBE_CONTAINER
					var group []string
					for _, value := range p.containers {
						group = append(group, value.ContainerId)
					}
					command.ContainerIds = group
					podCommand := message.PodCommand{
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
	p.compareAndSetStatus(POD_DELETED_STATUS)
	command := &message.CommandWithContainerIds{}
	command.CommandType = message.COMMAND_DELETE_CONTAINER
	var group []string
	for _, value := range p.containers {
		group = append(group, value.ContainerId)
	}
	command.ContainerIds = group
	podCommand := message.PodCommand{
		PodCommandType:   message.DELETE_POD,
		ContainerCommand: &(command.Command),
	}
	p.commandChan <- podCommand
	p.client.DeleteRuntimePod(p.GetName())
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
