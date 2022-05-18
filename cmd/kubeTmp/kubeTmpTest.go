package main

import (
	"fmt"
	"minik8s/pkg/etcdstore/nodeConfigStore"
)

func main() {
	nodeConfigStore.AddNewNode("11.23.45.67")
	nodeConfigStore.AddNewNode("177.29.34.22")
	res := nodeConfigStore.GetNodes()
	for _, v := range res {
		fmt.Println(v.MetaData.Name)
	}
}
