package service

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/listerwatcher"
	"sync"
	"time"
)

const (
	TimerInterval = 20
	PollCommand   = "Poll"
	NoPodsError   = "NoPodsError"
)

type RuntimeService struct {
	//service的配置文件
	serviceConfig *object.Service
	//service选择的Pod
	pods          []*object.Pod
	ls            *listerwatcher.ListerWatcher
	timerStopChan chan bool
	commandChan   chan command
	Client        client.RESTClient
	rwLock        sync.RWMutex
	//定时器用于轮询
	timer *time.Ticker
	//控制定时器请求是否接收，防止拥堵
	canPollSend bool
	Err         error
}
type command struct {
	CommandType string
}

//------------------------------------tools------------------------------------------//
func isExist(target string, from []string) bool {
	for _, val := range from {
		if target == val {
			return true
		}
	}
	return false
}
func (service *RuntimeService) selectPods(isInit bool) error {
	selector := service.serviceConfig.Spec.Selector
	res, err := service.ls.List(config.PodRuntimePrefix)
	if err != nil {
		return err
	}
	var origin []*object.Pod
	for _, val := range res {
		tmp := &object.Pod{}
		err = json.Unmarshal(val.ValueBytes, tmp)
		origin = append(origin, tmp)
	}
	//select pods
	var filter []*object.Pod
	for _, val := range origin {
		if val.Status.Phase != object.Running {
			continue
		}
		//考虑label，端口没开放是用户自己的问题，这里不管
		canChoose := true
		for k, v := range selector {
			podV, ok := val.Labels[k]
			if !ok {
				canChoose = false
				break
			}
			if v != podV {
				canChoose = false
				break
			}
		}
		if canChoose {
			filter = append(filter, val)
		}
	}
	//先把service里的坏的给去掉
	var okPods []*object.Pod
	for _, val := range service.pods {
		if val.Status.Phase == object.Running {
			okPods = append(okPods, val)
		}
	}
	service.pods = okPods
	//尝试填充,最多三个
	for _, val := range filter {
		if len(service.pods) >= 3 {
			//最多三个
			break
		}
		isExist := false
		for _, sPod := range service.pods {
			//已经存在
			if val.Name == sPod.Name {
				isExist = true
				break
			}
		}
		if isExist {
			continue
		}
		//不存在该pod，直接加入
		service.pods = append(service.pods, val)
	}
	//更新serviceConfig 并上传
	//如果没有选取到pod, 需要报错，同时如果已经是错误的不需要在去更新etcd
	var replace []object.PodNameAndIp
	updateEtcd := false
	if len(service.pods) == 0 {
		if len(service.serviceConfig.Spec.PodNameAndIps) != 0 {
			service.serviceConfig.Spec.PodNameAndIps = replace
			service.serviceConfig.Status.Phase = object.Failed
			service.serviceConfig.Status.Err = NoPodsError
			updateEtcd = true
		} else {
			if isInit {
				//第一次就没选到pod
				service.serviceConfig.Spec.PodNameAndIps = replace
				service.serviceConfig.Status.Phase = object.Failed
				service.serviceConfig.Status.Err = NoPodsError
			}
		}
	} else {
		//先生成replace
		for _, val := range service.pods {
			replace = append(replace, object.PodNameAndIp{Name: val.Name, Ip: val.Status.PodIP})
		}
		//更新serviceConfig
		service.serviceConfig.Spec.PodNameAndIps = replace
		service.serviceConfig.Status.Phase = object.Running
		if service.serviceConfig.Status.Err == NoPodsError {
			service.serviceConfig.Status.Err = ""
		}
		updateEtcd = true
	}
	//更新etcd
	if isInit {
		//第一次是一定要更新的
		err = service.Client.UpdateRuntimeService(service.serviceConfig)
		return err
	} else if updateEtcd {
		err = service.Client.UpdateRuntimeService(service.serviceConfig)
	}
	return err
}

//----------------------------------------------------------------------//

func NewRuntimeService(serviceConfig *object.Service, lsConfig *listerwatcher.Config, clientConfig client.Config) *RuntimeService {
	runtimeService := &RuntimeService{}
	runtimeService.commandChan = make(chan command, 100)
	runtimeService.serviceConfig = serviceConfig
	ls, _ := listerwatcher.NewListerWatcher(lsConfig)
	runtimeService.ls = ls
	runtimeService.Client = client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	var lock sync.RWMutex
	runtimeService.rwLock = lock
	runtimeService.canPollSend = false
	//启动syncLoop
	go runtimeService.syncLoop(runtimeService.commandChan)
	//selector
	runtimeService.Err = runtimeService.selectPods(true)
	//启动轮询
	runtimeService.canPollSend = true
	runtimeService.startPoll()
	return runtimeService
}
func (service *RuntimeService) syncLoop(commands <-chan command) {
	for {
		select {
		case cmd, ok := <-commands:
			if !ok {
				return
			}
			service.rwLock.Lock()
			switch cmd.CommandType {
			case PollCommand:
				//如果当前的pods为空，尝试select
				if len(service.pods) == 0 {
					service.Err = service.selectPods(false)
				} else {
					//检查pod状态， 如果有错调用select
					callSelect := false
					for _, pod := range service.pods {
						message, err := service.Client.GetRuntimePod(pod.Name)
						if err != nil {
							service.Err = err
							fmt.Println("[runtimeService] GetRuntimePod error")
							continue
						}
						if message == nil {
							//原pod已经被删除了
							pod.Status.Phase = object.Delete
							callSelect = true
							continue
						}
						if message.Status.Phase != object.Running {
							pod.Status.Phase = message.Status.Phase
							callSelect = true
						}
					}
					if callSelect {
						//调用selectPods
						service.Err = service.selectPods(false)
					}
				}
				service.canPollSend = true
			}
			service.rwLock.Unlock()
		}
	}
}
func (service *RuntimeService) startPoll() {
	service.timer = time.NewTicker(TimerInterval * time.Second)
	service.timerStopChan = make(chan bool)
	go func(service *RuntimeService) {
		defer service.timer.Stop()
		for {
			select {
			case <-service.timer.C:
				service.rwLock.Lock()
				if service.canPollSend {
					cmd := command{CommandType: PollCommand}
					service.commandChan <- cmd
					service.canPollSend = false
				}
				service.rwLock.Unlock()
			case <-service.timerStopChan:
				return
			}
		}
	}(service)
}
func (service *RuntimeService) DeleteService() {
	service.rwLock.Lock()
	defer service.rwLock.Unlock()
	//关闭定时器
	service.timerStopChan <- true
	//关闭隧道
	close(service.commandChan)
	//删除etcd中的东西
	err := service.Client.DeleteRuntimeService(service.serviceConfig.MetaData.Name)
	if err != nil {
		fmt.Println(err)
	}
}
