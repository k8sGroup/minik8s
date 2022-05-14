package client

import (
	"github.com/google/uuid"

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

func DefaultClientConfig() Config {
	return Config{
		Host: "127.0.0.1:8080",
	}
}

/******************************Pod*******************************/

func (r RESTClient) CreateRSPod(ctx context.Context, rs *object.ReplicaSet) error {
	id := uuid.New()

	podName := rs.Spec.Template.Name + id.String()
	attachURL := "/registry/pod/default/" + podName

	pod, _ := GetPodFromRS(rs)
	pod.Name = podName
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

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

func (r RESTClient) UpdatePods(pod *object.Pod) error {
	attachURL := "/registry/pod/default/" + pod.Name
	err := Put(r.Base+attachURL, pod)
	if err != nil {
		return err
	}
	return nil
}

func (r RESTClient) DeletePod(podName string) error {
	attachURL := "/registry/pod/default/" + podName
	err := Del(r.Base + attachURL)
	return err
}

// GetPodFromRS TODO: type conversion
func GetPodFromRS(rs *object.ReplicaSet) (*object.Pod, error) {
	pod := &object.Pod{}
	pod.Spec = rs.Spec.Template.Spec
	// add ownership
	owner := object.OwnerReference{
		Kind:       object.ReplicaSetKind,
		Name:       rs.Name,
		Controller: true,
	}
	pod.OwnerReferences = append(pod.OwnerReferences, owner)
	return pod, nil
}

func (r RESTClient) GetPod(name string) (*object.Pod, error) {
	attachUrl := "/registry/pod/default/" + name
	resp, err := Get(r.Base + attachUrl)
	if err != nil {
		return nil, err
	}
	result := &object.Pod{}
	err = json.Unmarshal(resp[0].ValueBytes, result)
	return result, err
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

func GetRSPods(ls *listerwatcher.ListerWatcher, name string) ([]*object.Pod, error) {
	raw, err := ls.List("/registry/pod/default")

	var pods []*object.Pod

	if len(raw) == 0 {
		return pods, nil
	}

	// unmarshal and filter by ownership
	for _, rawPod := range raw {
		pod := &object.Pod{}
		err = json.Unmarshal(rawPod.ValueBytes, &pod)
		if ownBy(pod.OwnerReferences, name) {
			pods = append(pods, pod)
		}
	}

	if err != nil {
		fmt.Printf("[GetRSPods] unmarshal fail\n")
	}

	return pods, nil
}

func (r RESTClient) DeleteRS(rsName string) error {
	attachURL := "/registry/rs/default/" + rsName
	err := Del(r.Base + attachURL)
	return err
}

func ownBy(ownerReferences []object.OwnerReference, owner string) bool {
	for _, ref := range ownerReferences {
		if ref.Name == owner {
			return true
		}
	}
	return false
}

func OwnByRs(pod *object.Pod) (bool, string) {
	ownerReferences := pod.OwnerReferences
	if len(ownerReferences) == 0 {
		return false, ""
	}

	// unmarshal and filter by ownership
	for _, owner := range ownerReferences {
		if owner.Kind == object.ReplicaSetKind {
			return true, owner.Name
		}
	}

	return false, ""
}

/********************************Node*****************************/

func GetNodes(ls *listerwatcher.ListerWatcher) ([]*object.Node, error) {
	raw, err := ls.List("/registry/node/default")
	if err != nil {
		fmt.Printf("[GetNodes] fail to get nodes\n")
	}

	var nodes []*object.Node

	if len(raw) == 0 {
		return nodes, nil
	}

	for _, rawNode := range raw {
		node := &object.Node{}
		err = json.Unmarshal(rawNode.ValueBytes, &node)
		nodes = append(nodes, node)
	}

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

/***********************************put****************************************/
func (r RESTClient) PutWrap(attachUrl string, requestBody any) error {
	if requestBody == nil {
		req, err2 := http.NewRequest("PUT", r.Base+attachUrl, nil)
		if err2 != nil {
			return err2
		}
		resp, err3 := http.DefaultClient.Do(req)
		if err3 != nil {
			return err3
		}
		if resp.StatusCode != 200 {
			s := fmt.Sprintf("http request error, attachUrl = %s, StatusCode = %d", attachUrl, resp.StatusCode)
			return errors.New(s)
		}
	} else {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		reqBody := bytes.NewBuffer(body)
		req, err2 := http.NewRequest("PUT", r.Base+attachUrl, reqBody)
		if err2 != nil {
			return err2
		}
		resp, err3 := http.DefaultClient.Do(req)
		if err3 != nil {
			return err3
		}
		if resp.StatusCode != 200 {
			s := fmt.Sprintf("http request error, attachUrl = %s, StatusCode = %d", attachUrl, resp.StatusCode)
			return errors.New(s)
		}
	}
	return nil
}
