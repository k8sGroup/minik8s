package netconfig

const BasicEthName = "ens3"
const DockerEthName = "docker0"
const IFaceRegex = "ens*|enp*"
const EtcdNetworkPrefix = "/registry/network/test"
const FlanneldPath = "/root/flannel/flanneld"
const MkDockerOptPath = "/root/flannel/mk-docker-opts.sh"
const masterIp = "10.119.11.164"
const EtcdEndPoint = "http://" + masterIp + ":2379"
const DockerServiceFilePath = "/usr/lib/systemd/system/docker.service"
const BASIC_MASK = "/16"
const NODE_MASK = "/24"

var GlobalIpMap = map[string]string{
	//子网ip到浮动ip的映射
	"192.168.1.4":  "10.119.11.159",
	"192.168.1.6":  "10.119.11.151",
	"192.168.1.10": "10.119.11.144",
	"192.168.1.7":  "10.119.11.164",
}
