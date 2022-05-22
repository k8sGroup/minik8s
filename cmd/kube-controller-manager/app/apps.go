package app

import (
	"context"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/pkg/controller/autoscaler"
	"minik8s/pkg/controller/deployment"
	"minik8s/pkg/controller/jobcontroller"
	"minik8s/pkg/controller/replicaset"
	"minik8s/pkg/klog"
)

func startReplicaSetController(ctx context.Context, controllerCtx util.ControllerContext) error {
	klog.Debugf("start running replicaset controller\n")
	go replicaset.NewReplicaSetController(ctx, controllerCtx).Run(ctx)
	return nil
}

func startDeploymentController(ctx context.Context, controllerCtx util.ControllerContext) error {
	klog.Debugf("start running deployment controller\n")
	deploymentController := deployment.NewDeploymentController(ctx, controllerCtx)
	go deploymentController.Run(ctx)
	return nil
}

func startAutoscalerController(ctx context.Context, controllerCtx util.ControllerContext) error {
	klog.Debugf("start running autoscaler controller\n")
	autoscalerController := autoscaler.NewAutoscalerController(ctx, controllerCtx)
	go autoscalerController.Run(ctx)
	return nil
}

func startJobController(ctx context.Context, controllerCtx util.ControllerContext) error {
	klog.Debugf("start running job controller\n")
	jobController := jobcontroller.NewJobController(ctx, controllerCtx)
	go jobController.Run(ctx)
	return nil
}
