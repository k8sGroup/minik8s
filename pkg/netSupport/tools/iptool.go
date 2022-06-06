package tools

import (
	"minik8s/pkg/netSupport/netconfig"
	"net"
	"strings"
)

//获取内网ip
func LocalIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

//比如, name可以为 ens33
func getIPv4ByInterface(name string) ([]string, error) {
	var ips []string

	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}
func GetEns3IPv4Addr() string {
	val, _ := getIPv4ByInterface(netconfig.BasicEthName)
	return val[0]
}
func GetDocker0IpAndMask() string {
	val, _ := getIPv4ByInterface(netconfig.DockerEthName)
	a, b, c, _ := GetFourField(val[0])
	res := a + "." + b + "." + c + ".1/24"
	return res
}
func GetDynamicIp() string {
	ips, _ := getIPv4ByInterface(netconfig.BasicEthName)
	return netconfig.GlobalIpMap[ips[0]]
}

func GetBasicIpAndMask(ipAndMask string) string {
	index := strings.Index(ipAndMask, ".")
	a := ipAndMask[:index]
	ipAndMask = ipAndMask[index+1:]
	index = strings.Index(ipAndMask, ".")
	b := ipAndMask[:index]
	return a + "." + b + ".0.0/16"
}

func GetFourField(ipV4 string) (string, string, string, string) {
	index := strings.Index(ipV4, ".")
	a := ipV4[:index]
	ipV4 = ipV4[index+1:]
	index = strings.Index(ipV4, ".")
	b := ipV4[:index]
	ipV4 = ipV4[index+1:]
	index = strings.Index(ipV4, ".")
	c := ipV4[:index]
	//ipV4 = ipV4[index+1:]
	//index = strings.Index(ipV4, "/")
	d := ipV4[index+1:]
	return a, b, c, d
}

//默认格式正确，不进行错误处理
func getIp(ipAndMask string) string {
	index := strings.Index(ipAndMask, "/")
	return ipAndMask[:index]
}

func getMask(ipAndMask string) string {
	index := strings.Index(ipAndMask, "/")
	return ipAndMask[index+1:]
}
