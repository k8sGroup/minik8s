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
		if selectType == scheduler.SelectRoundRobin {
			fmt.Println("user Round Robin Policy")
		}
		if selectType == scheduler.SelectAffinity {
			fmt.Println("use label match policy")
		}
	} else {
		fmt.Println("use default schedule police: random")
	}
	app.SchedulerRun(selectType)
}
