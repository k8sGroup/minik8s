package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	"net/http"
)

type Config struct {
	Host string // ip and port
}

type RESTClient struct {
	Base string // url = base+resource+name
}

/******************************Pod*******************************/

func (r RESTClient) CreatePods(ctx context.Context, template *object.PodTemplateSpec) error {
	pod, _ := GetPodFromTemplate(template)
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)
	attachURL := "/registry/pod/default"

	req, _ := http.NewRequest("PUT", r.Base+attachURL, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("create pod fail")
	}

	result := &object.Pod{}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail\n")
	}
	return err
}

func (r RESTClient) UpdatePods(ctx context.Context, pod *object.Pod) error {
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)
	attachURL := "/registry/pod/default/" + pod.Name

	req, _ := http.NewRequest("PUT", r.Base+attachURL, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("create pod fail")
	}
	defer resp.Body.Close()

	return nil
}

func (r RESTClient) DeletePod(ctx context.Context, podName string) error {
	attachURL := "/registry/pod/default" + podName
	req, _ := http.NewRequest("DELETE", r.Base+attachURL, nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("delete pod fail")
	}
	defer resp.Body.Close()

	return nil
}

// GetPodFromTemplate TODO: type conversion
func GetPodFromTemplate(template *object.PodTemplateSpec) (*object.Pod, error) {
	pod := &object.Pod{}
	pod.Spec = object.PodSpec{}
	return pod, nil
}

/********************************RS*****************************/

func (r RESTClient) GetRS(name string) (*object.ReplicaSet, error) {
	attachURL := "/registry/rs/default/" + name

	req, _ := http.NewRequest("GET", r.Base+attachURL, nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("delete pod fail")
	}

	result := &object.ReplicaSet{}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[GetRS] Body Pods Unmarshal fail\n")
	}
	return result, nil
}

func (r RESTClient) GetRSPods(name string) ([]*object.Pod, error) {
	attachURL := "/pod/" + name

	req, _ := http.NewRequest("GET", r.Base+attachURL, nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("get rs pod fail")
	}

	var result []*object.Pod
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, &result)
	if err != nil {
		klog.Infof("[GetRSPods] Body Pods Unmarshal fail\n")
	}

	return result, nil
}

func (r RESTClient) UpdateRSStatus(ctx context.Context, replicaSet *object.ReplicaSet) (*object.ReplicaSet, error) {
	body, _ := json.Marshal(replicaSet)
	reqBody := bytes.NewBuffer(body)
	attachURL := "/rs/" + replicaSet.Name + "/" + "status"

	req, _ := http.NewRequest("PUT", r.Base+attachURL, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("update rs fail")
	}

	result := &object.ReplicaSet{}
	body, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[UpdateRSStatus] Body Pods Unmarshal fail\n")
	}
	return result, nil
}

/********************************Node*****************************/

func GetNodes(ls *listerwatcher.ListerWatcher) ([]*object.Node, error) {
	raw, err := ls.List("/registry/node/default")
	if err != nil {
		fmt.Printf("[GetNodes] fail to get nodes\n")
	}

	fmt.Println(string(raw))

	var nodes []*object.Node
	err = json.Unmarshal(raw, &nodes)
	if err != nil {
		fmt.Printf("[GetNodes] unmarshal fail\n")
	}
	return nodes, nil
}

func (r RESTClient) RegisterNode(node *object.Node) error {
	if node.Name == "" {
		return errors.New("invalid node name")
	}
	attachURL := "/registry/node/default/" + node.Name
	body, err := json.Marshal(node)
	reqBody := bytes.NewBuffer(body)

	req, _ := http.NewRequest("PUT", r.Base+attachURL, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("[RegisterNode] unmarshal fail\n")
	}
	return nil
}

/********************************watch*****************************/

// WatchRegister get ticket for message queue
func (r RESTClient) WatchRegister(resource string, name string, withPrefix bool) (*string, *int64, error) {
	attachURL := "/" + resource + "/default"
	if !withPrefix {
		attachURL += "/" + name
	}
	req, _ := http.NewRequest("PUT", r.Base+attachURL, nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != 200 {
		return nil, nil, errors.New("api server register fail")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	result := &TicketResponse{}
	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail\n")
	}
	return &attachURL, &result.Ticket, nil
}
