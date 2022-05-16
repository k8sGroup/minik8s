package client

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"testing"
)

func TestUnmarshalProm(t *testing.T) {
	test := object.PromQueryRes{}
	str := "{\"status\":\"success\",\"data\":{\"resultType\":\"vector\",\"result\":[{\"metric\":{\"__name__\":\"node_monitor\",\"instance\":\"192.168.1.156:9070\",\"job\":\"node\",\"node\":\"node=test\",\"pod\":\"pod=\",\"resource\":\"cpu\"},\"value\":[1652587093.695,\"0.004550104166132792\"]}]}}"
	err := json.Unmarshal([]byte(str), &test)
	if err != nil {
		t.Errorf("unmarshal fail err:%v", err)
	}
	res := test.Data.ResultArray[0].Value[1]
	v := res.(string)
	fmt.Printf("res:%s\n", v)
	return
}

func TestGetResource(t *testing.T) {
	c := NewPromClient("http://localhost:9090")
	// get cpu
	cpuPercent, err := c.GetResource(object.CPU_RESOURCE, "", nil)
	if err != nil || cpuPercent == nil {
		t.Fail()
	}
	fmt.Printf("cpu:%f\n", *cpuPercent)

	// get memory
	memPercent, err := c.GetResource(object.MEMORY_RESOURCE, "", nil)
	if err != nil {
		t.Fail()
	}
	fmt.Printf("cpu:%f\n", *memPercent)
}
