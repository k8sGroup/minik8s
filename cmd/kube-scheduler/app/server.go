package app

import (
	"context"
	"minik8s/pkg/client"
	"minik8s/pkg/messaging"
	"minik8s/pkg/scheduler"
	"time"
)

func SchedulerRun() {
	msgConfig := messaging.QConfig{
		User:          "guest",
		Password:      "guest",
		Host:          "127.0.0.1",
		Port:          "5672",
		MaxRetry:      5,
		RetryInterval: time.Duration(1000),
	}
	clientConfig := client.Config{
		Host: "127.0.0.1:8080",
	}
	sched := scheduler.NewScheduler(msgConfig, clientConfig)
	sched.Run(context.TODO())
	select {}
}
