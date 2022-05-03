package kubeproxy

import (
	"encoding/json"
	"errors"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/etcdstore/netConfigStore"
	"minik8s/pkg/klog"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/kubeproxy/commandWorker"
	"minik8s/pkg/kubeproxy/iptablesManager"
	"minik8s/pkg/kubeproxy/netconfig"
	"minik8s/pkg/kubeproxy/tools"
	"minik8s/pkg/listerwatcher"
	"sync"
)

//--------------常量定义---------------------//
const OP_ADD_GRE = 1
const OP_DELETE_GRE = 2 //不需要考虑delete的情况
const OP_BOOT_NET = 3
const NOT_BOOT = 1
const BOOT_FAILED = 2
const IS_BOOTING = 3
const BOOT_SUCCEED = 4
const MASK = "need allocate port"

//------------------------------------------//
type Kubeproxy struct {
	portMappings []iptablesManager.PortMapping
	rwLock       sync.RWMutex
	//存一份map,由clusterIp映射到管道名, 同时这个与etcd中的数据做对比可以得到要增加或者删除的gre端口
	ipPipeMap map[string]string
	//存一份snapshoop
	kubeproxySnapShoot     KubeproxySnapShoot
	ls                     *listerwatcher.ListerWatcher
	Client                 client.RESTClient
	stopChannel            <-chan struct{}
	netCommandChan         chan NetCommand
	NetCommandResponseChan chan NetCommandResponse
	commandWorker          *commandWorker.CommandWorker
	//自己所在节点的ip
	myClusterIp string
	//自己分配的网段
	myIpAndMask string
	bootStatus  int
	err         error
}

type NetCommand struct {
	//操作类型
	Op        int
	ClusterIp string
	//这些只有boot的时候需要, 即分配给自己机子的网段，同时通过这个可以得到大网段
	IpAndMask string
}
type NetCommandResponse struct {
	//操作类型
	Op        int
	ClusterIp string
	GrePort   string
	//Err 为nil代表command正确执行
	Err error
}
type KubeproxySnapShoot struct {
	PortMappings []iptablesManager.PortMapping
	IpPipeMap    map[string]string
	IpPairs      []netConfigStore.IpPair
	MyClusterIp  string
	MyIpAndMask  string
	Error        string
	BootStatus   int
}

func NewKubeproxy(lsConfig *listerwatcher.Config, clientConfig client.Config) (*Kubeproxy, error) {
	newKubeproxy := &Kubeproxy{}
	var rwLock sync.RWMutex
	newKubeproxy.rwLock = rwLock
	newKubeproxy.ipPipeMap = make(map[string]string)
	newKubeproxy.stopChannel = make(chan struct{}, 10)
	newKubeproxy.netCommandChan = make(chan NetCommand, 100)
	newKubeproxy.NetCommandResponseChan = make(chan NetCommandResponse, 100)
	newKubeproxy.bootStatus = NOT_BOOT
	newKubeproxy.commandWorker = &commandWorker.CommandWorker{}
	ls, err2 := listerwatcher.NewListerWatcher(lsConfig)
	if err2 != nil {
		return nil, err2
	}
	newKubeproxy.ls = ls
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newKubeproxy.Client = restClient
	ips, err := tools.GetIPv4ByInterface(netconfig.ETH_NAME)
	if err != nil {
		newKubeproxy.err = err
	} else {
		newKubeproxy.myClusterIp = ips[0]
	}
	newKubeproxy.kubeproxySnapShoot = KubeproxySnapShoot{
		PortMappings: newKubeproxy.portMappings,
		IpPipeMap:    newKubeproxy.ipPipeMap,
		MyClusterIp:  newKubeproxy.myClusterIp,
		MyIpAndMask:  newKubeproxy.myIpAndMask,
		Error:        newKubeproxy.err.Error(),
		BootStatus:   newKubeproxy.bootStatus,
	}
	return newKubeproxy, nil
}
func (k *Kubeproxy) StartKubeProxy() {
	//先启动线程
	go k.commandWorker.SyncLoop(k.netCommandChan, k.NetCommandResponseChan)
	go k.listeningResponse()
	//注册
	k.registerNode()
}
func (k *Kubeproxy) registerNode() error {
	//先挂上watch
	go k.ls.Watch(config.NODE_PREFIX, k.watchRegister, k.stopChannel)

	//获取所有其他的节点
	res, err := k.getIpPairs()
	if err != nil {
		return err
	}
	//设置map， 暂时不生成command，用MASK占位
	for _, value := range res {
		k.ipPipeMap[value.ClusterIp] = MASK
	}
	//发起注册的http请求
	attachURL := "/node/register/" + k.myClusterIp
	err = k.Client.PutWrap(attachURL, nil)
	if err != nil {
		return err
	}
	return nil
}
func (k *Kubeproxy) getIpPairs() ([]netConfigStore.IpPair, error) {
	raw, err := k.ls.List(config.NODE_PREFIX)
	if err != nil {
		return nil, err
	}
	var res []netConfigStore.IpPair
	if len(raw) == 0 {
		return res, nil
	}
	for _, rawPair := range raw {
		ipPair := &netConfigStore.IpPair{}
		err = json.Unmarshal(rawPair.ValueBytes, &ipPair)
		res = append(res, *ipPair)
	}
	return res, nil
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

//-------------------------------注册相关----------------------------------------//
func (kp *Kubeproxy) watchRegister(res etcdstore.WatchRes) {
	//只管增加不管删除
	if res.ResType == etcdstore.DELETE {
		//不需要管其他节点删除的情况
		return
	}
	ipPair := &netConfigStore.IpPair{}
	err := json.Unmarshal(res.ValueBytes, ipPair)
	if err != nil {
		klog.Warnf("watchRegister unmarshal faield")
	}
	kp.watchAndHandleInner(ipPair)
}

//只有先init才能去进行其他gre端口的创建
//所以如果没有boot同时有一大堆gre端口等着创建，可以先在map中生成空映射，等到收到boot成功的response再去生成command
//此外不需要考虑gre端口的销毁，因为每个gre端口是唯一生成的，不会重复，另外就算其他节点没注册但是连接着,网段的唯一性也可以保证不会出错

func (k *Kubeproxy) watchAndHandleInner(ipPair *netConfigStore.IpPair) {
	k.rwLock.Lock()
	defer k.rwLock.Unlock()
	switch k.bootStatus {
	case NOT_BOOT:
		//根据是否是自己的ip区别对待
		if ipPair.ClusterIp == k.myClusterIp {
			command := NetCommand{
				Op:        OP_BOOT_NET,
				ClusterIp: ipPair.ClusterIp,
				IpAndMask: ipPair.NodeIpAndMask,
			}
			k.myIpAndMask = ipPair.NodeIpAndMask
			k.bootStatus = IS_BOOTING
			k.netCommandChan <- command
		} else {
			k.ipPipeMap[ipPair.ClusterIp] = MASK
		}
		return
	case BOOT_FAILED:
		//上次boot失败，尝试再次boot
		k.ipPipeMap[ipPair.ClusterIp] = MASK
		command := NetCommand{
			Op:        OP_BOOT_NET,
			ClusterIp: k.myClusterIp,
			IpAndMask: k.myIpAndMask,
		}
		k.netCommandChan <- command
		return
	case IS_BOOTING:
		//同样不需要生成实际的command
		k.ipPipeMap[ipPair.ClusterIp] = MASK
		return
	case BOOT_SUCCEED:
		//生成command
		command := NetCommand{
			Op:        OP_ADD_GRE,
			ClusterIp: ipPair.ClusterIp,
		}
		k.netCommandChan <- command
		return
	}
}

func (k *Kubeproxy) listeningResponse() {
	for {
		select {
		case response, ok := <-k.NetCommandResponseChan:
			if !ok {
				return
			}
			switch response.Op {
			case OP_ADD_GRE:
				k.rwLock.Lock()
				if response.Err == nil {
					//success
					k.ipPipeMap[response.ClusterIp] = response.GrePort
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
								Op:        OP_ADD_GRE,
								ClusterIp: key,
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
