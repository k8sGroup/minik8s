package config

import (
	"github.com/spf13/pflag"
	"minik8s/cmd/kube-controller-manager/app/controllers"
)

type ControllerOptions interface {
	AddFlags(fs *pflag.FlagSet)
	SetDefault()
}

type KubeControllerManagerOptions struct {
	// TODO : add more controllers here
	ReplicaSetController *controllers.ReplicaSetControllerOptions
	DeploymentController *controllers.DeploymentControllerOptions
	AutoscalerController *controllers.AutoscalerControllerOptions
}

func NewKubeControllerManagerOptions() *KubeControllerManagerOptions {
	controllerManagerOptions := KubeControllerManagerOptions{
		&controllers.ReplicaSetControllerOptions{},
		&controllers.DeploymentControllerOptions{},
		&controllers.AutoscalerControllerOptions{},
	}
	controllerManagerOptions.SetDefault()
	return &controllerManagerOptions
}

func (opts *KubeControllerManagerOptions) Flags() *pflag.FlagSet {
	// TODO : add more flags
	flagSet := pflag.FlagSet{}
	addFlags(opts.ReplicaSetController, &flagSet)
	addFlags(opts.DeploymentController, &flagSet)
	addFlags(opts.AutoscalerController, &flagSet)
	return &flagSet
}

func (opts *KubeControllerManagerOptions) SetDefault() {
	// TODO
	setDefault(opts.ReplicaSetController)
	setDefault(opts.DeploymentController)
	setDefault(opts.AutoscalerController)
}

func (opts *KubeControllerManagerOptions) Config() *Config {
	// TODO : finish this function
	return &Config{
		ReplicaSetControllerOptions: opts.ReplicaSetController,
		DeploymentControllerOptions: opts.DeploymentController,
		AutoscalerControllerOptions: opts.AutoscalerController,
	}
}

func addFlags(controllerOptions ControllerOptions, flagSet *pflag.FlagSet) {
	controllerOptions.AddFlags(flagSet)
}

func setDefault(controllerOptions ControllerOptions) {
	controllerOptions.SetDefault()
}
