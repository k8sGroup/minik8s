package app

import (
	"context"
	"minik8s/pkg/controller/replicaset"
)

func startReplicaSetController(ctx context.Context, controllerCtx ControllerContext) (bool, error) {
	go replicaset.NewReplicaSetController().Run(ctx)
	return true, nil
}
