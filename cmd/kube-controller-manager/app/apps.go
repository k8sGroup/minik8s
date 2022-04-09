package app

import (
	"context"
	"fmt"
	"time"
)


func startReplicaSetController(ctx context.Context, controllerContext ControllerContext) (controller.Interface, bool, error) {
	// go replicaset.NewReplicaSetController(
	// 	controllerContext.InformerFactory.Apps().V1().ReplicaSets(),
	// 	controllerContext.InformerFactory.Core().V1().Pods(),
	// 	controllerContext.ClientBuilder.ClientOrDie("replicaset-controller"),
	// 	replicaset.BurstReplicas,
	// ).Run(ctx, int(controllerContext.ComponentConfig.ReplicaSetController.ConcurrentRSSyncs))
	return nil, true, nil
}