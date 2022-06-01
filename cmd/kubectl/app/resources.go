package app

import set "github.com/deckarep/golang-set/v2"

// 用于命令行参数校验
// 不是yaml文件中的kind！！！
//var commandLineResources set.Set[string]
var commandLineResource set.Set[string]

func init() {
	//commandLineResources = set.NewSet[string]("pods", "deployments", "replicasets", "svc","service")
	commandLineResource = set.NewSet[string]("pod", "deployment", "replicaset", "service", "dns", "autoscaler")
}
