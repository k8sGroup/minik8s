package config

import (
	"minik8s/cmd/kube-controller-manager/app/controllers"
)

type Config struct {
	*controllers.ReplicaSetControllerOptions
	*controllers.DeploymentControllerOptions
	*controllers.AutoscalerControllerOptions
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() *CompletedConfig {
	// TODO : complete podConfig
	return &CompletedConfig{c}
}
