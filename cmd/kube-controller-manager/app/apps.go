package app

import (
	"context"
	"minik8s/pkg/controller/replicaset"
	"minik8s/pkg/klog"
)

func startReplicaSetController(ctx context.Context, controllerCtx ControllerContext) error {
	klog.Debugf("start running replicaset controller\n")
	go replicaset.NewReplicaSetController(ctx, controllerCtx).Run(ctx)
	return nil
}
