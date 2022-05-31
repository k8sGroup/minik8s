package app

import set "github.com/deckarep/golang-set/v2"

// 用于命令行参数校验
// 不是yaml文件中的kind！！！
var commandLineResources set.Set[string]
var commandLineResource set.Set[string]
var plural2singular map[string]string

func init() {
	commandLineResources = set.NewSet[string]("pods", "deployments", "replicasets", "svc")
	commandLineResource = set.NewSet[string]("pod", "deployment", "replicaset", "svc")
	plural2singular = map[string]string{
		"pods":        "pod",
		"deployments": "deployment",
		"replicasets": "replicaset",
		"svc":         "svc",
	}
}
