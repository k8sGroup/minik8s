package kubeNetSupport

import (
	"encoding/json"
	"errors"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/kubeNetSupport/boot"
	"minik8s/pkg/kubeNetSupport/iptablesManager"
	"minik8s/pkg/kubeNetSupport/netconfig"
	"minik8s/pkg/kubeNetSupport/tools"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/listerwatcher"
	"sync"
	"time"
)

//--------------常量定义---------------------//
const OP_ADD_VXLAN = 1
const OP_DELETE_GRE = 2 //不需要考虑delete的情况
const OP_BOOT_NET = 3
const NOT_BOOT = 1
const BOOT_FAILED = 2
const IS_BOOTING = 3
const BOOT_SUCCEED = 4
const MASK = "need allocate port"

//------------------------------------------//
type KubeNetSupport struct {
	portMappings []iptablesManager.PortMapping
	rwLock       sync.RWMutex
	//存一份map,由physicalIp映射到管道名, 同时这个与etcd中的数据做对比可以得到要增加或者删除的gre端口
	ipPipeMap map[string]string
	//存一份snapshoop2
	kubeproxySnapShoot     KubeNetSupportSnapShoot
	ls                     *listerwatcher.ListerWatcher
	Client                 client.RESTClient
	stopChannel            <-chan struct{}
	netCommandChan         chan NetCommand
	NetCommandResponseChan chan NetCommandResponse
	commandWorker          *CommandWorker
	//自己所在节点的ip
	myphysicalIp string
	//自己分配的网段
	myIpAndMask string
	//节点的名字
	myNodeName string
	bootStatus int
	err        error
}
type CommandWorker struct {
	isMaster bool
}

func (worker *CommandWorker) SyncLoop(commands <-chan NetCommand, responses chan<- NetCommandResponse) {
	for {
		select {
		case command, ok := <-commands:
			if !ok {
				return
			}
			switch command.Op {
			case OP_ADD_VXLAN:
				portName := netconfig.FormVxLanPort()
				err := boot.SetVxLanPortInBr0(portName, command.physicalIp)
				response := NetCommandResponse{
					Op:         command.Op,
					physicalIp: command.physicalIp,
					VxLanPort:  portName,
					Err:        err,
				}
				responses <- response
			case OP_BOOT_NET:
				err := boot.BootNetWork(command.IpAndMask, tools.GetBasicIpAndMask(command.IpAndMask), worker.isMaster)
				response := NetCommandResponse{
					Op:  command.Op,
					Err: err,
				}
				responses <- response
			}
		}
	}
}

type NetCommand struct {
	//操作类型
	Op         int
	physicalIp string
	//这些只有boot的时候需要, 即分配给自己机子的网段，同时通过这个可以得到大网段
	IpAndMask string
}
type NetCommandResponse struct {
	//操作类型
	Op         int
	physicalIp string
	VxLanPort  string
	//Err 为nil代表command正确执行
	Err error
}
type KubeNetSupportSnapShoot struct {
	PortMappings []iptablesManager.PortMapping
	//map ip to pipe
	IpPipeMap    map[string]string
	MyphysicalIp string
	MyIpAndMask  string
	NodeName     string
	Error        string
	BootStatus   int
}

func NewKubeNetSupport(lsConfig *listerwatcher.Config, clientConfig client.Config, isMaster bool) (*KubeNetSupport, error) {
	newKubeNetSupport := &KubeNetSupport{}
	var rwLock sync.RWMutex
	newKubeNetSupport.rwLock = rwLock
	newKubeNetSupport.ipPipeMap = make(map[string]string)
	newKubeNetSupport.stopChannel = make(chan struct{}, 10)
	newKubeNetSupport.netCommandChan = make(chan NetCommand, 100)
	newKubeNetSupport.NetCommandResponseChan = make(chan NetCommandResponse, 100)
	newKubeNetSupport.bootStatus = NOT_BOOT
	newKubeNetSupport.commandWorker = &CommandWorker{
		isMaster: isMaster,
	}
	ls, err2 := listerwatcher.NewListerWatcher(lsConfig)
	if err2 != nil {
		return nil, err2
	}
	newKubeNetSupport.ls = ls
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newKubeNetSupport.Client = restClient
	newKubeNetSupport.myphysicalIp = tools.GetEns3IPv4Addr()
	sErr := ""
	if newKubeNetSupport.err != nil {
		sErr = newKubeNetSupport.err.Error()
	}
	newKubeNetSupport.kubeproxySnapShoot = KubeNetSupportSnapShoot{
		PortMappings: newKubeNetSupport.portMappings,
		IpPipeMap:    newKubeNetSupport.ipPipeMap,
		MyphysicalIp: newKubeNetSupport.myphysicalIp,
		MyIpAndMask:  newKubeNetSupport.myIpAndMask,
		Error:        sErr,
		BootStatus:   newKubeNetSupport.bootStatus,
		NodeName:     newKubeNetSupport.myNodeName,
	}
	return newKubeNetSupport, nil
}
func (k *KubeNetSupport) StartKubeNetSupport() error {
	//先启动线程
	go k.commandWorker.SyncLoop(k.netCommandChan, k.NetCommandResponseChan)
	go k.listeningResponse()
	//注册
	fmt.Println("start register")
	return k.registerNode()
}
func (k *KubeNetSupport) registerNode() error {
	//先挂上watch
	go k.ls.Watch(config.NODE_PREFIX, k.watchRegister, k.stopChannel)
	time.Sleep(2 * time.Second)
	//获取所有其他的节点
	res, err := k.getNodes()
	if err != nil {
		return err
	}
	fmt.Println("ipPairs get is\n")
	fmt.Println(res)
	//设置map， 暂时不生成command，用MASK占位
	for _, value := range res {
		k.ipPipeMap[value.Spec.PhysicalIp] = MASK
	}
	//发起注册的http请求
	fmt.Println(k.myphysicalIp)
	attachURL := config.NODE_PREFIX + "/" + k.myphysicalIp
	err = k.Client.PutWrap(attachURL, nil)
	if err != nil {
		return err
	}
	return nil
}
func (k *KubeNetSupport) getNodes() ([]object.Node, error) {
	raw, err := k.ls.List(config.NODE_PREFIX)
	if err != nil {
		return nil, err
	}
	var res []object.Node
	if len(raw) == 0 {
		return res, nil
	}
	for _, rawPair := range raw {
		node := &object.Node{}
		err = json.Unmarshal(rawPair.ValueBytes, node)
		res = append(res, *node)
	}
	return res, nil
}
func (k *KubeNetSupport) RemovePortMapping(pod *pod.PodSnapShoot, dockerPort string, hostPort string) error {
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
func (kubeproxy *KubeNetSupport) AddPortMapping(pod *pod.PodSnapShoot, dockerPort string, hostPort string) (iptablesManager.PortMapping, error) {
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
func (kubeproxy *KubeNetSupport) GetKubeproxySnapShoot() KubeNetSupportSnapShoot {
	if kubeproxy.rwLock.TryRLock() {
		sErr := ""
		if kubeproxy.err != nil {
			sErr = kubeproxy.err.Error()
		}
		kubeproxy.kubeproxySnapShoot = KubeNetSupportSnapShoot{
			PortMappings: kubeproxy.portMappings,
			IpPipeMap:    kubeproxy.ipPipeMap,
			MyphysicalIp: kubeproxy.myphysicalIp,
			MyIpAndMask:  kubeproxy.myIpAndMask,
			Error:        sErr,
			BootStatus:   kubeproxy.bootStatus,
			NodeName:     kubeproxy.myNodeName,
		}
		kubeproxy.rwLock.RUnlock()
		return kubeproxy.kubeproxySnapShoot
	} else {
		return kubeproxy.kubeproxySnapShoot
	}
}

//-------------------------------注册相关----------------------------------------//
func (kp *KubeNetSupport) watchRegister(res etcdstore.WatchRes) {
	//只管增加不管删除
	if res.ResType == etcdstore.DELETE {
		//不需要管其他节点删除的情况
		return
	}
	node := &object.Node{}
	err := json.Unmarshal(res.ValueBytes, node)
	if err != nil {
		klog.Warnf("watchRegister unmarshal faield")
	}
	kp.watchAndHandleInner(node)
}

//只有先init才能去进行其他gre端口的创建
//所以如果没有boot同时有一大堆gre端口等着创建，可以先在map中生成空映射，等到收到boot成功的response再去生成command
//此外不需要考虑gre端口的销毁，因为每个gre端口是唯一生成的，不会重复，另外就算其他节点没注册但是连接着,网段的唯一性也可以保证不会出错

func (k *KubeNetSupport) watchAndHandleInner(node *object.Node) {
	k.rwLock.Lock()
	defer k.rwLock.Unlock()
	fmt.Println("watchAndHandle receive command\n")
	switch k.bootStatus {
	case NOT_BOOT:
		//根据是否是自己的ip区别对待
		if node.Spec.PhysicalIp == k.myphysicalIp {
			command := NetCommand{
				Op:         OP_BOOT_NET,
				physicalIp: node.Spec.PhysicalIp,
				IpAndMask:  node.Spec.NodeIpAndMask,
			}
			k.myIpAndMask = node.Spec.NodeIpAndMask
			k.myNodeName = node.MetaData.Name
			k.bootStatus = IS_BOOTING
			k.netCommandChan <- command
		} else {
			k.ipPipeMap[node.Spec.PhysicalIp] = MASK
		}
		return
	case BOOT_FAILED:
		//上次boot失败，尝试再次boot
		k.ipPipeMap[node.Spec.PhysicalIp] = MASK
		command := NetCommand{
			Op:         OP_BOOT_NET,
			physicalIp: k.myphysicalIp,
			IpAndMask:  k.myIpAndMask,
		}
		k.netCommandChan <- command
		return
	case IS_BOOTING:
		//同样不需要生成实际的command
		k.ipPipeMap[node.Spec.PhysicalIp] = MASK
		return
	case BOOT_SUCCEED:
		//生成command
		command := NetCommand{
			Op:         OP_ADD_VXLAN,
			physicalIp: node.Spec.PhysicalIp,
		}
		k.netCommandChan <- command
		return
	}
}

func (k *KubeNetSupport) listeningResponse() {
	for {
		select {
		case response, ok := <-k.NetCommandResponseChan:
			if !ok {
				return
			}
			switch response.Op {
			case OP_ADD_VXLAN:
				k.rwLock.Lock()
				if response.Err == nil {
					//success
					k.ipPipeMap[response.physicalIp] = response.VxLanPort
				} else {
					k.err = response.Err
					//直接设置error, 同时不选择重试，因为这时候基本重试也会寄
				}
				k.rwLock.Unlock()
			case OP_BOOT_NET:
				k.rwLock.Lock()
				if response.Err == nil {
					k.bootStatus = BOOT_SUCCEED
					//处理之前没有分配grePort的node
					for key, v := range k.ipPipeMap {
						if v == MASK {
							command := NetCommand{
								Op:         OP_ADD_VXLAN,
								physicalIp: key,
							}
							k.netCommandChan <- command
						}
					}
				} else {
					//设置状态为boot failed，且只有watch再次来新的时候才会重试
					k.bootStatus = BOOT_FAILED
					k.err = response.Err
				}
				k.rwLock.Unlock()
			}
		}
	}
}

//---------------------------------------------------------------------//
