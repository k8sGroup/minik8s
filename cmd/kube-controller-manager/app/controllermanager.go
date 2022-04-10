package app

import "context"

type Controller interface {
}

type ControllerContext struct {
}

type InitFunc func(ctx context.Context, controllerCtx ControllerContext) (enabled bool, err error)

func NewControllerInitializers() map[string]InitFunc {
	controller := map[string]InitFunc{}
	// TODO : Initialize the map with controller name and InitFunc
	return controller
}

func StartControllers(ctx context.Context, controllerContext ControllerContext, controllers map[string]InitFunc) {
	for controllerName, initFunc := range controllers {
		enabled, err := initFunc(ctx, controllerContext)
		if err != nil {
			return
		}
	}
}
