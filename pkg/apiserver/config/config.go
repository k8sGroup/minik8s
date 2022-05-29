package config

import (
	"minik8s/pkg/messaging"
	"time"
)

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

const SharedDataDirectory = "/home/SharedData"

const Path = "/registry/:resource/:namespace/:resourceName"
const PrefixPath = "/registry/:resource/:namespace"
const ParamResource = "resource"
const ParamResourceName = "resourceName"
const ParamType = "type"
const NODE_NAME = "name"

// UserPath is the API only for user operation
// key is the path cutting user, and add uuid
const UserPodPath = "/user/registry/pod/:namespace/:resourceName"
const UserRSPath = "/user/registry/rs/:namespace/:resourceName"

// path for kube client
const (
	RS               = "/registry/rs/default/:resourceName"
	PodRuntime       = "/registry/pod/default/:resourceName"
	PodRuntimePrefix = "/registry/pod/default"
	NODE             = "/registry/node/default/:resourceName"
	NODE_PREFIX      = "/registry/node/default"

	PodCONFIG       = "/registry/podConfig/default/:resourceName"
	PodConfigPREFIX = "/registry/podConfig/default"

	ServiceConfig       = "/registry/serviceConfig/default/:resourceName"
	ServiceConfigPrefix = "/registry/serviceConfig/default"
	Service             = "/registry/service/default/:resourceName"
	ServicePrefix       = "/registry/service/default"

	RSConfig       = "/registry/rsConfig/default/:resourceName"
	RSConfigPrefix = "/registry/rsConfig/default"

	SharedData       = "/registry/sharedData/default/:resourceName"
	SharedDataPrefix = "/registry/sharedData/default"

	DnsAndTrans       = "/registry/dnsAndTrans/default/:resourceName"
	DnsAndTransPrefix = "/registry/dnsAndTrans/default"
	RS_POD = "/rs/pod"
)

var defaultValidResources = []string{"pod", "rs", "deployment", "node", "test", "autoscaler", "podConfig", "sharedData", "service", "job", "serviceConfig", "rsConfig", "dnsAndTrans"}

type ServerConfig struct {
	HttpPort       int
	ValidResources []string // 合法的resource
	EtcdEndpoints  []string // etcd集群每一个节点的ip和端口
	EtcdTimeout    time.Duration
	QueueConfig    *messaging.QConfig
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		HttpPort:       8080,
		ValidResources: defaultValidResources,
		EtcdEndpoints:  []string{"localhost:12379"},
		EtcdTimeout:    5 * time.Second,
		QueueConfig:    messaging.DefaultQConfig(),
	}
}
