package main

import (
	"fmt"
	"minik8s/cmd/kubelet/app/pod"
)

func main() {
	fmt.Printf(pod.GetCurrentAbPathByCaller())
}
