package app

import (
	"context"
	"minik8s/cmd/kube-controller-manager/app/config"
	"minik8s/cmd/kube-controller-manager/app/options"
	"minik8s/pkg/klog"

	"minik8s/pkg/listerwatcher"

	"github.com/spf13/cobra"
)

type Informer struct {
}

type Controller interface {
}

type ControllerContext struct {
	ls *listerwatcher.ListerWatcher
}

type InitFunc func(ctx context.Context, controllerCtx ControllerContext) (err error)

func NewControllerManagerCommand() *cobra.Command {
	opts, err := options.NewKubeControllerManagerOptions()
	if err != nil {
		klog.Fatalf("failed to initialize kube controller manager options\n")
	}
	cmd := &cobra.Command{
		Use: "kube-controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			c, err := opts.Config()
			if err != nil {
				klog.Fatalf("failed to configure controller manager %v\n", err)
			}
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
	controllerContext, err := CreateControllerContext()
	if err != nil {
		return err
	}
	if err := StartControllers(context.TODO(), controllerContext, NewControllerInitializers()); err != nil {
		klog.Fatalf("error starting controllers: %v\n", err)
	}
	// TODO
	select {}

	return nil
}

func CreateControllerContext() (*ControllerContext, error) {
	ls, err := listerwatcher.NewListerWatcher(listerwatcher.DefaultConfig())
	if err != nil {
		return nil, err
	}
	controllerContext := &ControllerContext{ls: ls}
	return controllerContext, nil
}

func NewControllerInitializers() map[string]InitFunc {
	controller := map[string]InitFunc{}
	// TODO : Initialize the map with controller name and InitFunc
	controller["replicaset"] = startReplicaSetController
	return controller
}

func StartControllers(ctx context.Context, controllerContext *ControllerContext, controllers map[string]InitFunc) error {
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
