package commandWorker

import (
	"minik8s/pkg/kubeproxy"
	"minik8s/pkg/kubeproxy/boot"
	"minik8s/pkg/kubeproxy/netconfig"
	"minik8s/pkg/kubeproxy/tools"
)

type CommandWorker struct {
}

func (worker *CommandWorker) SyncLoop(commands <-chan kubeproxy.NetCommand, responses chan<- kubeproxy.NetCommandResponse) {
	for {
		select {
		case command, ok := <-commands:
			if !ok {
				return
			}
			switch command.Op {
			case kubeproxy.OP_ADD_GRE:
				portName := netconfig.FormGrePort()
				err := boot.SetGrePortInBr0(portName, command.ClusterIp)
				response := kubeproxy.NetCommandResponse{
					Op:        command.Op,
					ClusterIp: command.ClusterIp,
					GrePort:   portName,
					Err:       err,
				}
				responses <- response
			case kubeproxy.OP_BOOT_NET:
				err := boot.BootNetWork(command.IpAndMask, tools.GetBasicIpAndMask(command.IpAndMask))
				response := kubeproxy.NetCommandResponse{
					Op:  command.Op,
					Err: err,
				}
				responses <- response
			}
		}
	}
}
