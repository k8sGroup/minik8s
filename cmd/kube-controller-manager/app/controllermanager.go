package app

import "context"

type ControllerContext struct {
}

type InitFunc func(ctx context.Context, controllerCtx ControllerContext) (enabled bool, err error)

func NewControllerInitializers() map[string]InitFunc {
	controllers := map[string]InitFunc{}
	controllers["replicaset"] = startReplicaSetController
	return controllers
}
