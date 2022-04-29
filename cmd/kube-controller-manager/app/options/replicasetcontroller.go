package options

import (
	"github.com/spf13/pflag"
)

type ReplicaSetControllerOptions struct {
	ConcurrentRSSyncs int32
}

func (o *ReplicaSetControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.Int32Var(&o.ConcurrentRSSyncs, "concurrent-replicaset-syncs", o.ConcurrentRSSyncs,
		"The number of replica sets that are allowed to sync concurrently. "+
			"Larger number = more responsive replica management, but more CPU (and network) load")
}

func (o *ReplicaSetControllerOptions) SetDefault() {
	// TODO
}
