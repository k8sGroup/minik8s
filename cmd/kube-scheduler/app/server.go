package app

import (
	"context"
	"minik8s/pkg/client"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/scheduler"
)

func SchedulerRun() {
	clientConfig := client.Config{Host: "127.0.0.1:8080"}
	sched := scheduler.NewScheduler(listerwatcher.DefaultConfig(), clientConfig)
	sched.Run(context.TODO())
	select {}
}
