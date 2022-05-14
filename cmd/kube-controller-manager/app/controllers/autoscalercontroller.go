package controllers

import "github.com/spf13/pflag"

type AutoscalerControllerOptions struct {
	ScaleIntervals int
}

func (o *AutoscalerControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IntVar(&o.ScaleIntervals, "scaler-interval", o.ScaleIntervals,
		"Interval in seconds of autoscaler controller's checking.")
}

func (o *AutoscalerControllerOptions) SetDefault() {
	o.ScaleIntervals = 10
}
