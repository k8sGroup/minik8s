package app

import (
	"context"
	"github.com/spf13/cobra"
	"minik8s/cmd/kube-controller-manager/app/config"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
)

type InitFunc func(ctx context.Context, controllerCtx util.ControllerContext) (err error)

func NewControllerManagerCommand() *cobra.Command {
	opts := config.NewKubeControllerManagerOptions()
	cmd := &cobra.Command{
		Use: "kube-controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			c := opts.Config()
			// FIXME : what's the meaning of stopCh ?
			if err := Run(c.Complete()); err != nil {
				klog.Fatalf("failed to run controller manager%v\n", err)
			}
		},
	}
	cmd.Flags().AddFlagSet(opts.Flags())
	return cmd
}

func Run(c *config.CompletedConfig) error {
	controllerContext, err := CreateControllerContext(c)
	if err != nil {
		return err
	}
	if err := StartControllers(context.TODO(), controllerContext, NewControllerInitializers()); err != nil {
		klog.Fatalf("error starting controllers: %v\n", err)
	}
	// TODO: give each controller a new unique ls
	select {}
}

// CreateControllerContext TODO: make global podConfig variable
func CreateControllerContext(c *config.CompletedConfig) (*util.ControllerContext, error) {
	ls, err := listerwatcher.NewListerWatcher(listerwatcher.DefaultConfig())
	if err != nil {
		return nil, err
	}
	controllerContext := &util.ControllerContext{
		Ls:             ls,
		MasterIP:       "127.0.0.1",
		HttpServerPort: "8080",
		PromServerPort: "9090",
		Config:         c,
	}
	return controllerContext, nil
}

func NewControllerInitializers() map[string]InitFunc {
	controller := map[string]InitFunc{}
	// TODO : Initialize the map with controller name and InitFunc
	controller["replicaset"] = startReplicaSetController
	controller["deployment"] = startDeploymentController
	controller["autoscaler"] = startAutoscalerController
	controller["job"] = startJobController
	controller["service"] = startServiceController
	return controller
}

func StartControllers(ctx context.Context, controllerContext *util.ControllerContext, controllers map[string]InitFunc) error {
	for controllerName, initFunc := range controllers {
		klog.Infof("Starting controller %s\n", controllerName)
		err := initFunc(ctx, *controllerContext)
		if err != nil {
			klog.Errorf("Error starting %s\n", controllerName)
			return err
		}
		klog.Debugf("Started controller %s\n", controllerName)
	}
	return nil
}
