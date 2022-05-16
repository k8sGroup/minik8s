package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"minik8s/object"
	"net/http"
	"strconv"
)

type PromClient struct {
	Base string
}

// NewPromClient base:http://localhost:9090
func NewPromClient(base string) *PromClient {
	return &PromClient{
		Base: base,
	}
}

// GetResource  only support get cpu/memory used percentage of a pod
// podName, podUUID
func (c *PromClient) GetResource(resource object.Resource, podName string, podUID string, tagPair *string) (*float64, error) {
	// make query url
	url, err := c.makeQuery(resource, podName, podUID, tagPair)
	if err != nil || url == nil {
		return nil, err
	}

	// do request
	request, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	val, err := getPromRespValue(data)
	return val, err
}

func (c *PromClient) makeQuery(resource object.Resource, podName string, podUID string, tagPair *string) (*string, error) {
	var resourceTag string
	if resource == object.CPU_RESOURCE {
		resourceTag = "cpu"
	} else if resource == object.MEMORY_RESOURCE {
		resourceTag = "memory"
	} else {
		return nil, errors.New("invalid resource")
	}
	var query string
	// query=node_monitor{resource="cpu"}
	if tagPair == nil {
		query = fmt.Sprintf("query=node_monitor{resource=\"%s\",pod=\"%s\",uid=\"%s\"}", resourceTag, podName, podUID)
	} else {
		query = fmt.Sprintf("query=node_monitor{resource=\"%s\",pod=\"%s\",uid=\"%s\",selector=\"%s\"}", resourceTag, podName, podUID, *tagPair)
	}
	url := c.Base + "/api/v1/query?" + query
	return &url, nil
}

func getPromRespValue(data []byte) (*float64, error) {
	prom := object.PromQueryRes{}
	err := json.Unmarshal(data, &prom)
	if err != nil {
		fmt.Printf("[getPromRespValue] unmarshal fail err:%v\n", err)
		return nil, errors.New("unmarshal fail err")
	}
	if len(prom.Data.ResultArray) == 0 || len(prom.Data.ResultArray[0].Value) < 2 {
		fmt.Printf("[getPromRespValue] bad query message\n")
		return nil, errors.New("bad query message")
	}
	res := prom.Data.ResultArray[0].Value[1]
	v := res.(string)
	resFloat, err := strconv.ParseFloat(v, 64)
	return &resFloat, nil
}
