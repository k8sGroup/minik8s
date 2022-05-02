package deployment

import (
	"context"
	"minik8s/cmd/kube-controller-manager/app"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"time"
)

type DeploymentController struct {
	ls          *listerwatcher.ListerWatcher
	cm          *concurrentmap.ConcurrentMapTrait[int, int]
	stopChannel chan struct{}
}

func NewDeploymentController(ctx context.Context, controllerCtx app.ControllerContext) *DeploymentController {
	dc := &DeploymentController{
		ls:          controllerCtx.GetListerWatcher(),
		cm:          concurrentmap.NewConcurrentMapTrait[int, int](),
		stopChannel: make(chan struct{}),
	}
	return dc
}

func (dc *DeploymentController) Run(ctx context.Context) {
	klog.Debugf("[DeploymentController] running...\n")

	<-ctx.Done()
	close(dc.stopChannel)
}

func (dc *DeploymentController) register() {
	go func() {
		for {
			err := dc.ls.Watch("/registry/deployment", dc.addDeployment, dc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/deployment\n")
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (dc *DeploymentController) addDeployment(res etcdstore.WatchRes) {

}

func (dc *DeploymentController) deleteDeployment(res etcdstore.WatchRes) {

}

func (dc *DeploymentController) deletePod(res etcdstore.WatchRes) {

}
