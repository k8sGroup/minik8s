package main

import "minik8s/pkg/mesh"

func main() {
	p := mesh.NewProxy("172.16.24.2")
	p.Init()
	p.Run()
}
