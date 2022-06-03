package mesh_sidecar

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Sidecar struct {
	ls          *listerwatcher.ListerWatcher
	stopChannel <-chan struct{}
}

func NewSidecar(lsConfig *listerwatcher.Config) *Sidecar {
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Println("[NewRouter] list watch fail...")
	}

	return &Sidecar{
		ls: ls,
	}
}

func (d *Sidecar) Run() {
	rand.Seed(time.Now().Unix())
	klog.Debugf("[ReplicaSetController]start running\n")
	go d.register()
	select {}
}

func (d *Sidecar) register() {
	watchSidecar := func(d *Sidecar) {
		err := d.ls.Watch(config.Sidecar, d.watchSidecar, d.stopChannel)
		if err != nil {
			fmt.Printf("[Router] ListWatch init fail...")
		}
	}

	go watchSidecar(d)
}

func (d *Sidecar) watchSidecar(res etcdstore.WatchRes) {
	sidecar := &object.SidecarInject{}
	err := json.Unmarshal(res.ValueBytes, sidecar)
	if err != nil {
		fmt.Println("[watchSidecar] Unmarshall fail")
		return
	}

	inbound := sidecar.Inbound
	outbound := sidecar.Outbound
	SysUid := sidecar.SysUid
	if SysUid != 1337 {
		fmt.Printf("[watchSidecar] bad sysuid:%v, should be 1337\n", SysUid)
		return
	}

	if sidecar.Status == true {
		err := InitSidecar(inbound, outbound, SysUid)
		if err != nil {
			fmt.Printf("[watchSidecar] init iptables err:%v\n", err)
		}
	} else {
		err := FinalizeSidecar()
		if err != nil {
			fmt.Printf("[watchSidecar] init iptables err:%v\n", err)
		}
	}
}

// InitSidecar inbound 15006 outbound 15001 uid 1337
func InitSidecar(inbound int, outbound int, sysUID int) error {
	out, err := execCmdWithOutput("docker", fmt.Sprintf("run --rm --name=istio-init --network=host --cap-add=NET_ADMIN istio/proxyv2 istio-iptables -p %v -z %v -u %v -m REDIRECT -i * -b * -d 15020", outbound, inbound, sysUID))
	if err != nil {
		fmt.Printf("[InitSidecar] error:%v out:%v\n", err, out)
		fmt.Println("[InitSidecar] Reset iptables...")
		out, err = execCmdWithOutput("docker", "run --rm --name=istio-init --network=host --cap-add=NET_ADMIN istio/proxyv2 istio-clean-iptables")
		if err != nil {
			fmt.Printf("[InitSidecar] Reset fail, error:%v out:%v\n", err, out)
			return err
		}
		out, err = execCmdWithOutput("docker", fmt.Sprintf("run --rm --name=istio-init --network=host --cap-add=NET_ADMIN istio/proxyv2 istio-iptables -p %v -z %v -u %v -m REDIRECT -i * -b * -d 15020", outbound, inbound, sysUID))
		return err
	}
	return err
}

func FinalizeSidecar() error {
	out, err := execCmdWithOutput("docker", "run --rm --name=istio-init --network=host --cap-add=NET_ADMIN istio/proxyv2 istio-clean-iptables")
	if err != nil {
		fmt.Printf("[FinalizeSidecar] error:%v out:%v\n", err, out)
	}
	return err
}

func execCmdWithOutput(exc string, args string) ([]string, error) {
	cmd := exec.Command(exc, strings.Split(args, " ")...)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(stdout)
	var result []string
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		result = append(result, line)
	}
	err = cmd.Wait()
	return result, err
}
