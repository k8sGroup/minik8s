package netConfigStore

import (
	"fmt"
	"strings"
	"sync"
)

//这个是大网段，确保为node分配的网段全在这里边
const BASIC_IP_AND_MASK = "172.17.0.0/16"
const BASIC_MASK = "/16"
const NODE_MASK = "/24"
const NODE_LAST_FIELD = "1"

var count = 1
var lock sync.Mutex

func GetNodeIpAndMask() string {
	lock.Lock()
	defer lock.Unlock()
	a, b, _, _ := getFourField(BASIC_IP_AND_MASK)
	res := a + "." + b + "." + fmt.Sprintf("%d", count) + "." + NODE_LAST_FIELD + "" + NODE_MASK
	count++
	return res
}

//----------------------tools begin-----------------------------//
//默认格式正确，不进行错误处理
func getIp(ipAndMask string) string {
	index := strings.Index(ipAndMask, "/")
	return ipAndMask[:index]
}

func getMask(ipAndMask string) string {
	index := strings.Index(ipAndMask, "/")
	return ipAndMask[index+1:]
}
func getFourField(ipAndMask string) (string, string, string, string) {
	index := strings.Index(ipAndMask, ".")
	a := ipAndMask[:index]
	ipAndMask = ipAndMask[index+1:]
	index = strings.Index(ipAndMask, ".")
	b := ipAndMask[:index]
	ipAndMask = ipAndMask[index+1:]
	index = strings.Index(ipAndMask, ".")
	c := ipAndMask[:index]
	ipAndMask = ipAndMask[index+1:]
	index = strings.Index(ipAndMask, "/")
	d := ipAndMask[:index]
	return a, b, c, d
}

//---------------------tools end--------------------------------//
