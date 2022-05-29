package client

import (
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcdstore"

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
	podUID := uuid.New().String()
	attachURL := config.PodConfigPREFIX + "/" + rs.Spec.Template.Name + podUID

	pod, _ := GetPodFromRS(rs)
	pod.Name = rs.Spec.Template.Name + podUID
	pod.UID = podUID
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)

	// put config
	req, _ := http.NewRequest("PUT", r.Base+attachURL, reqBody)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("create pod config fail")
	}

	//// put runtime
	//attachURL = "/registry/pod/default/" + rs.Spec.Template.Name + podUID
	//req, _ = http.NewRequest("PUT", r.Base+attachURL, reqBody)
	//resp, _ = http.DefaultClient.Do(req)

	//if resp.StatusCode != object.SUCCESS {
	//	return errors.New("create pod fail")
	//}

	return nil
}

func (r RESTClient) UpdateRuntimePod(pod *object.Pod) error {
	attachURL := "/registry/pod/default/" + pod.Name
	err := Put(r.Base+attachURL, pod)
	if err != nil {
		return err
	}
	return nil
}

func (r RESTClient) DeleteRuntimePod(podName string) error {
	attachURL := "/registry/pod/default/" + podName
	err := Del(r.Base + attachURL)
	return err
}
func (r RESTClient) UpdateConfigPod(pod *object.Pod) error {
	attachURL := config.PodConfigPREFIX + "/" + pod.Name
	err := Put(r.Base+attachURL, pod)
	return err
}
func (r RESTClient) DeleteConfigPod(podName string) error {
	attachURL := config.PodConfigPREFIX + "/" + podName
	err := Del(r.Base + attachURL)
	return err
}

func (r RESTClient) GetConfigPod(name string) (*object.Pod, error) {
	attachUrl := config.PodConfigPREFIX + "/" + name
	resp, err := Get(r.Base + attachUrl)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, nil
	}
	result := &object.Pod{}
	err = json.Unmarshal(resp[0].ValueBytes, result)
	return result, err
}

// GetPodFromRS TODO: type conversion
func GetPodFromRS(rs *object.ReplicaSet) (*object.Pod, error) {
	pod := &object.Pod{}
	pod.Spec = rs.Spec.Template.Spec
	// add ownership
	owner := object.OwnerReference{
		Kind:       object.ReplicaSetKind,
		Name:       rs.Name,
		UID:        rs.UID,
		Controller: true,
	}
	pod.OwnerReferences = append(pod.OwnerReferences, owner)
	return pod, nil
}

func (r RESTClient) GetRuntimePod(name string) (*object.Pod, error) {
	attachUrl := "/registry/pod/default/" + name
	resp, err := Get(r.Base + attachUrl)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, nil
	}
	result := &object.Pod{}
	err = json.Unmarshal(resp[0].ValueBytes, result)
	return result, err
}

/********************************RS*****************************/

func GetRuntimeRS(ls *listerwatcher.ListerWatcher, name string) (*object.ReplicaSet, error) {
	attachURL := "/registry/rs/default/" + name

	raw, err := ls.List(attachURL)
	if err != nil {
		fmt.Printf("[GetRS] fail to get nodes\n")
	}

	if len(raw) == 0 {
		return nil, errors.New("not find")
	}

	result := &object.ReplicaSet{}

	err = json.Unmarshal(raw[0].ValueBytes, result)

	if err != nil {
		fmt.Printf("[GetRS] unmarshal fail\n")
	}
	fmt.Printf("[GetRS] rs:%+v\n", result)
	return result, nil
}

func GetRSPods(ls *listerwatcher.ListerWatcher, name string, UID string) ([]*object.Pod, error) {
	raw, err := ls.List("/registry/pod/default")
	if err != nil {
		fmt.Printf("[GetRSPods] list fail\n")
		return nil, err
	}
	return MakePods(raw, name, UID)
}

func MakePods(raw []etcdstore.ListRes, name string, UID string) ([]*object.Pod, error) {
	var pods []*object.Pod

	if len(raw) == 0 {
		return pods, nil
	}

	// unmarshal and filter by ownership
	for _, rawPod := range raw {
		pod := &object.Pod{}
		err := json.Unmarshal(rawPod.ValueBytes, &pod)
		if err != nil {
			fmt.Printf("[GetRSPods] unmarshal fail\n")
			return nil, err
		}
		if ownBy(pod.OwnerReferences, name, UID) {
			pods = append(pods, pod)
		}
	}

	return pods, nil
}

func (r RESTClient) DeleteRS(rsName string) error {
	attachURL := "/registry/rs/default/" + rsName
	fmt.Printf("delete rs:" + attachURL + "\n")
	err := Del(r.Base + attachURL)
	return err
}

func ownBy(ownerReferences []object.OwnerReference, owner string, UID string) bool {
	for _, ref := range ownerReferences {
		if ref.Name == owner && ref.UID == UID {
			return true
		}
	}
	return false
}

func OwnByRs(pod *object.Pod) (exist bool, name string, UID string) {
	ownerReferences := pod.OwnerReferences
	if len(ownerReferences) == 0 {
		return false, "", ""
	}

	// unmarshal and filter by ownership
	for _, owner := range ownerReferences {
		if owner.Kind == object.ReplicaSetKind {
			fmt.Printf("[OwnByRs] owner:%v\n", owner.Name)
			return true, owner.Name, owner.UID
		}
	}

	return false, "", ""
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

/*******************************Service**********************************/
func (r RESTClient) UpdateService(service *object.Service) error {
	attachUrl := config.ServiceConfigPrefix + "/" + service.MetaData.Name
	err := Put(r.Base+attachUrl, service)
	return err
}
func (r RESTClient) UpdateRuntimeService(service *object.Service) error {
	attachUrl := config.ServicePrefix + "/" + service.MetaData.Name
	err := Put(r.Base+attachUrl, service)
	return err
}
func (r RESTClient) GetRuntimeService(name string) (*object.Service, error) {
	attachUrl := config.ServicePrefix + "/" + name
	resp, err := Get(r.Base + attachUrl)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, nil
	}
	result := &object.Service{}
	err = json.Unmarshal(resp[0].ValueBytes, result)
	return result, err
}
func (r RESTClient) DeleteRuntimeService(name string) error {
	attachUrl := config.ServicePrefix + "/" + name
	err := Del(r.Base + attachUrl)
	return err
}

/***************************DnsAndTrans************************************/
func (r RESTClient) UpdateDnsAndTrans(trans *object.DnsAndTrans) error {
	attachUrl := config.DnsAndTransPrefix + "/" + trans.MetaData.Name
	err := Put(r.Base+attachUrl, trans)
	return err
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
