package app

import set "github.com/deckarep/golang-set/v2"

// 用于命令行参数校验
// 不是yaml文件中的kind！！！
//var commandLineResources set.Set[string]
var commandLineResource set.Set[string]
var plural2singular map[string]string

func init() {
	//commandLineResources = set.NewSet[string]("pods", "deployments", "replicasets", "svc","service")
	commandLineResource = set.NewSet[string]("pod", "deployment", "replicaset", "service", "dns")
	plural2singular = map[string]string{
		"pods":        "pod",
		"deployments": "deployment",
		"replicasets": "replicaset",
		"svc":         "svc",
	}
}
