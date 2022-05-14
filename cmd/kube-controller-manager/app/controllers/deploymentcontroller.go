package controllers

import "github.com/spf13/pflag"

type DeploymentControllerOptions struct {
	ResyncIntervals int
}

func (o *DeploymentControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IntVar(&o.ResyncIntervals, "metadata-resync", o.ResyncIntervals,
		"Interval in seconds of deployment controller's re-synchronization.")
}

func (o *DeploymentControllerOptions) SetDefault() {
	o.ResyncIntervals = 15
}
