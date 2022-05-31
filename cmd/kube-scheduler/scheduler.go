package main

import (
	"fmt"
	"minik8s/cmd/kube-scheduler/app"
	"minik8s/pkg/scheduler"
	"os"
)

func main() {
	selectType := scheduler.SelectRandom
	if len(os.Args) != 1 {
		selectType = os.Args[1]
		if selectType != scheduler.SelectRandom && selectType != scheduler.SelectRoundRobin && selectType != scheduler.SelectAffinity {
			fmt.Printf("unKnown type:%s, use default schedule police: random", selectType)
			selectType = scheduler.SelectRandom
		}
	} else {
		fmt.Println("use default schedule police: random")
	}
	app.SchedulerRun(selectType)
}
