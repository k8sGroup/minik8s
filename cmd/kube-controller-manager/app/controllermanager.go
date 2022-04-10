package app

import (
	"context"
	"minik8s/pkg/klog"
)

type Controller interface {
}

type ControllerContext struct {
}

type InitFunc func(ctx context.Context, controllerCtx ControllerContext) (err error)

func CreateControllerContext() (ControllerContext, error) {
	controllerContext := ControllerContext{}
	return controllerContext, nil
}

func NewControllerInitializers() map[string]InitFunc {
	controller := map[string]InitFunc{}
	// TODO : Initialize the map with controller name and InitFunc
	return controller
}

func StartControllers(ctx context.Context, controllerContext ControllerContext, controllers map[string]InitFunc) error {
	for controllerName, initFunc := range controllers {
		klog.Infof("Starting controller %s\n", controllerName)
		err := initFunc(ctx, controllerContext)
		if err != nil {
			klog.Errorf("Error starting %s\n", controllerName)
			return err
		}
		klog.Infof("Started controller %s\n", controllerName)
	}
	return nil
}
