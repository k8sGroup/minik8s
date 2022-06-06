package main

import (
	"fmt"
	"minik8s/cmd/kube-scheduler/app"
	"minik8s/pkg/scheduler"
	"os"
)

func main() {
	selectType := scheduler.SelectRandom
	if len(os.Args) > 1 {
		selectType = os.Args[1]
		switch selectType {
		case "random":
			selectType = scheduler.SelectRandom
		case scheduler.SelectRandom:
			selectType = scheduler.SelectRandom
		case "round-robin":
			selectType = scheduler.SelectRoundRobin
		case scheduler.SelectRoundRobin:
			selectType = scheduler.SelectRoundRobin
		case "affinity":
			selectType = scheduler.SelectAffinity
		case scheduler.SelectAffinity:
			selectType = scheduler.SelectAffinity
		default:
			fmt.Printf("unknown type:%s, use default schedule police: random", selectType)
			selectType = scheduler.SelectRandom
		}
	} else {
		fmt.Println("use default schedule police: random")
	}
	app.SchedulerRun(selectType)
}
