package options

import (
	"github.com/spf13/pflag"
	"minik8s/cmd/kube-controller-manager/app/config"
)

type ControllerOptions interface {
	AddFlags(fs *pflag.FlagSet)
	SetDefault()
}

type KubeControllerManagerOptions struct {
	// TODO : add more options here
	ReplicaSetController *ReplicaSetControllerOptions
}

func NewKubeControllerManagerOptions() (*KubeControllerManagerOptions, error) {
	options := KubeControllerManagerOptions{}
	options.SetDefault()
	return &options, nil
}

func (opts *KubeControllerManagerOptions) Flags() *pflag.FlagSet {
	// TODO : add more flags
	flagSet := pflag.FlagSet{}
	addFlags(opts.ReplicaSetController, &flagSet)
	return &flagSet
}

func (opts *KubeControllerManagerOptions) SetDefault() {
	// TODO
	setDefault(opts.ReplicaSetController)
}

func (opts *KubeControllerManagerOptions) Config() (*config.Config, error) {
	// TODO : finish this function
	return &config.Config{}, nil
}

func addFlags(controllerOptions ControllerOptions, flagSet *pflag.FlagSet) {
	controllerOptions.AddFlags(flagSet)
}

func setDefault(controllerOptions ControllerOptions) {
	controllerOptions.SetDefault()
}
