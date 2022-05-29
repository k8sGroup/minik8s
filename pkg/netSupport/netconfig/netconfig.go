package netconfig

const BasicEthName = "ens3"
const DockerEthName = "docker0"
const IFaceRegex = "ens*|enp*"
const EtcdNetworkPrefix = "/registry/network/test"

//文件路径

const FlanneldPath = "/root/flannel/flanneld"
const MkDockerOptPath = "/root/flannel/mk-docker-opts.sh"
const NginxDirModelPath = "/root/nginx/nginxModule"
const MasterIp = "10.119.11.164"
const EtcdEndPoint = "http://" + MasterIp + ":2379"
const DockerServiceFilePath = "/usr/lib/systemd/system/docker.service"

//DNS 与 网关相关
const (
	GateWayRsModulePath      string = "/home/minik8s/build/buildRs/gateWayRs.yaml"
	CoreDnsRsModulePath      string = "/home/minik8s/build/buildRs/coreDnsRs.yaml"
	GateWayServiceModulePath string = "/home/minik8s/build/buildService/gateWayService.yaml"
	CoreDnsServiceModulePath string = "/home/minik8s/build/buildService/coreDnsService.yaml"
	GateWayRsNamePrefix      string = "gateWayRs"
	GateWayPodNamePrefix     string = "gateWayPod"
	GateWayContainerPrefix   string = "gateWayContainer"
	GateWayServicePrefix     string = "gateWayService"
	NginxPathPrefix          string = "/root/nginx"
	NginxConfigFileName      string = "nginx.conf"
	CoreDnsConfigPath        string = "/root/coredns/hostsfile"
	BelongKey                string = "belong"
)

//一些服务的地址

const ServiceDns = "10.10.10.10"

//网关对应的gateWay中的nginx容器部分名字

var GlobalIpMap = map[string]string{
	//子网ip到浮动ip的映射
	"192.168.1.4":  "10.119.11.159",
	"192.168.1.6":  "10.119.11.151",
	"192.168.1.10": "10.119.11.144",
	"192.168.1.7":  "10.119.11.164",
}
