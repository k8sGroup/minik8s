package config

import "time"

/*
apiServer 的路径是 /registry/{resource}/{namespace}/{resource_name}

# deployment
/registry/deployments/default/game-2048
/registry/deployments/kube-system/prometheus-operator

# replicasets
/registry/replicasets/default/game-2048-c7d589ccf

# pod
/registry/pods/default/game-2048-c7d589ccf-8lsbw

# statefulsets
/registry/statefulsets/kube-system/prometheus-k8s

# daemonsets
/registry/daemonsets/kube-system/kube-proxy

# secrets
/registry/secrets/default/default-token-tbfmb

# serviceaccounts
/registry/serviceaccounts/default/default
*/

const Path = "/registry/:resource/:namespace/:resourceName"
const PrefixPath = "/registry/:resource/:namespace"
const ParamResource = "resource"

type Config struct {
	HttpPort       int
	ValidResources []string // 合法的resource
	EtcdEndpoints  []string // etcd集群每一个节点的ip和端口
	EtcdTimeout    time.Duration
}

func DefaultServerConfig() *Config {
	return &Config{
		HttpPort:       8080,
		ValidResources: []string{"pod"},
		EtcdEndpoints:  []string{"localhost:2379"},
		EtcdTimeout:    5 * time.Second,
	}
}