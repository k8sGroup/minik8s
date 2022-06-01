package netSupport

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport/boot"
	"minik8s/pkg/netSupport/netconfig"
	"minik8s/pkg/netSupport/tools"
	"sync"
	"time"
)

//------------------------------------------//
type KubeNetSupport struct {
	rwLock sync.RWMutex
	//存一份snapshoop2
	kubeproxySnapShoot KubeNetSupportSnapShoot
	node               *object.Node
	ls                 *listerwatcher.ListerWatcher
	Client             client.RESTClient
	stopChannel        <-chan struct{}
	//浮动ip
	myDynamicIp string
	//节点的名称
	myNodeName string
	//节点的docker网段
	myIpAndMask string
	err         error
}

type KubeNetSupportSnapShoot struct {
	MyDynamicIp string
	MyIpAndMask string
	NodeName    string
	Error       string
}

func NewKubeNetSupport(lsConfig *listerwatcher.Config, clientConfig client.Config, node *object.Node) (*KubeNetSupport, error) {
	newKubeNetSupport := &KubeNetSupport{}
	var rwLock sync.RWMutex
	newKubeNetSupport.rwLock = rwLock
	newKubeNetSupport.stopChannel = make(chan struct{}, 10)
	newKubeNetSupport.myDynamicIp = tools.GetDynamicIp()
	newKubeNetSupport.myIpAndMask = tools.GetDocker0IpAndMask()
	newKubeNetSupport.node = node
	ls, err2 := listerwatcher.NewListerWatcher(lsConfig)
	if err2 != nil {
		return nil, err2
	}
	newKubeNetSupport.ls = ls
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	newKubeNetSupport.Client = restClient
	sErr := ""
	if newKubeNetSupport.err != nil {
		sErr = newKubeNetSupport.err.Error()
	}
	newKubeNetSupport.kubeproxySnapShoot = KubeNetSupportSnapShoot{
		MyIpAndMask: newKubeNetSupport.myIpAndMask,
		MyDynamicIp: newKubeNetSupport.myDynamicIp,
		NodeName:    newKubeNetSupport.myNodeName,
		Error:       sErr,
	}
	return newKubeNetSupport, nil
}
func (k *KubeNetSupport) StartKubeNetSupport() error {
	fmt.Println("start register")
	return k.registerNode()
}
func (k *KubeNetSupport) registry() {
	netSupportRegister := func() {
		for {
			err := k.ls.Watch(config.NODE_PREFIX, k.watchRegister, k.stopChannel)
			if err != nil {
				fmt.Println("[kubeNetSupport] watch register error" + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}
	go netSupportRegister()
}
func (k *KubeNetSupport) registerNode() error {
	//先挂上watch
	k.registry()
	fmt.Println("start init flannel, please wait")
	boot.BootFlannel()
	//发起注册的http请求
	attachURL := config.NODE_PREFIX + "/" + k.myDynamicIp
	var node *object.Node
	if k.node == nil {
		node = &object.Node{
			MasterIp: netconfig.MasterIp,
			Spec: object.NodeSpec{
				DynamicIp:     k.myDynamicIp,
				NodeIpAndMask: k.myIpAndMask,
			},
		}
	} else {
		node = &object.Node{
			MetaData: k.node.MetaData,
			MasterIp: k.node.MasterIp,
			Spec: object.NodeSpec{
				DynamicIp:     k.myDynamicIp,
				NodeIpAndMask: k.myIpAndMask,
			},
		}
	}
	err := k.Client.PutWrap(attachURL, node)
	if err != nil {
		return err
	}
	return nil
}

func (netSupport *KubeNetSupport) GetKubeproxySnapShoot() KubeNetSupportSnapShoot {
	if netSupport.rwLock.TryRLock() {
		sErr := ""
		if netSupport.err != nil {
			sErr = netSupport.err.Error()
		}
		netSupport.kubeproxySnapShoot = KubeNetSupportSnapShoot{
			NodeName:    netSupport.myNodeName,
			MyDynamicIp: netSupport.myDynamicIp,
			MyIpAndMask: netSupport.myIpAndMask,
			Error:       sErr,
		}
		netSupport.rwLock.RUnlock()
		return netSupport.kubeproxySnapShoot
	} else {
		return netSupport.kubeproxySnapShoot
	}
}

//-------------------------------注册相关----------------------------------------//
func (kp *KubeNetSupport) watchRegister(res etcdstore.WatchRes) {
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

func (k *KubeNetSupport) watchAndHandleInner(node *object.Node) {
	k.rwLock.Lock()
	defer k.rwLock.Unlock()
	fmt.Println("watchAndHandle receive command\n")
	if node.Spec.DynamicIp == k.myDynamicIp {
		k.myNodeName = node.MetaData.Name
	}
	return
}

//---------------------------------------------------------------------//
