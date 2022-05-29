package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/controller"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/etcdstore/serviceConfigStore"
	"net/http"
	"strings"
)

// do not delete pod in etcd directly, just modify the status
//对pod的删除通过修改pod配置文件里的phase为DELETED进行
func (s *Server) deletePod(ctx *gin.Context) {
	name := ctx.Param(config.ParamResourceName)
	key := config.PodConfigPREFIX + "/" + name
	resList, err := s.store.Get(key)
	if err != nil || len(resList) == 0 {
		fmt.Printf("[deletePod] pod not exist:%s\n", name)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	pod := object.Pod{}
	err = json.Unmarshal(resList[0].ValueBytes, &pod)
	if err != nil {
		fmt.Printf("[deletePod] pod unmarshal fail\n")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	pod.Status.Phase = object.Delete
	raw, _ := json.Marshal(pod)
	err = s.store.Put(key, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

//同上述对pod的删除
func (s *Server) deleteService(ctx *gin.Context) {
	name := ctx.Param(config.ParamResourceName)
	key := config.ServiceConfigPrefix + "/" + name
	resList, err := s.store.Get(key)
	if err != nil || len(resList) == 0 {
		fmt.Printf("[deleteService] service not exist:%s\n", name)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	service := &object.Service{}
	err = json.Unmarshal(resList[0].ValueBytes, service)
	if err != nil {
		fmt.Println("[deleteService] service unmarshall fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	service.Status.Phase = object.Delete
	raw, _ := json.Marshal(service)
	err = s.store.Put(key, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

// do not delete rs in etcd directly, just modify the number of replicas
// TODO: real deletion by replica set controller !
func (s *Server) deleteRS(ctx *gin.Context) {
	name := ctx.Param(config.ParamResourceName)
	key := config.RSConfigPrefix + "/" + name
	resList, err := s.store.Get(key)
	if err != nil || len(resList) == 0 {
		fmt.Printf("[deleteRS] rs not exist:%s\n", name)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	rs := object.ReplicaSet{}
	err = json.Unmarshal(resList[0].ValueBytes, &rs)
	if err != nil {
		fmt.Printf("[deleteRS] pod unmarshal fail\n")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// if already zero just delete
	if rs.Spec.Replicas == 0 {
		fmt.Printf("[deleteRS] real del key %v\n", key)
		err = s.store.Del(key)
		ctx.Status(http.StatusOK)
		return
	}

	// set spec replicas to zero
	rs.Spec.Replicas = 0
	raw, _ := json.Marshal(rs)
	fmt.Printf("[deleteRS] put key %v\n", key)
	err = s.store.Put(key, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

// just for user to do some operation
// user add pod has unique name as the key, but also need to make uuid for other use
func (s *Server) userAddPod(ctx *gin.Context) {
	key := strings.TrimPrefix(ctx.Request.URL.Path, "/user")
	uid := uuid.New().String()
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := object.Pod{}
	err = json.Unmarshal(body, &pod)
	pod.UID = uid
	if err != nil {
		fmt.Printf("[deletePod] pod unmarshal fail\n")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(pod)
	fmt.Printf("key:%v\n", key)

	err = s.store.Put(key, body)
}

// just for user to do some operation
func (s *Server) userAddRS(ctx *gin.Context) {
	uid := uuid.New().String()
	body, err := ioutil.ReadAll(ctx.Request.Body)
	rs := object.ReplicaSet{}
	err = json.Unmarshal(body, &rs)
	rs.UID = uid
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	body, _ = json.Marshal(rs)
	err = s.store.Put(config.RSConfigPrefix+"/"+rs.Name, body)
}

//service part
func (s *Server) AddService(ctx *gin.Context) {
	//做service缺省值的填充处理
	body, err := ioutil.ReadAll(ctx.Request.Body)
	service := &object.Service{}
	err = json.Unmarshal(body, service)
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if service.Spec.Type == "" {
		service.Spec.Type = object.ClusterIp
	}
	ok, ip := serviceConfigStore.JudgeAndAllocClusterIp(service.MetaData.Name, service.Spec.ClusterIp)
	if !ok {
		fmt.Println("[AddService] ClusterIp illegal")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	service.Spec.ClusterIp = ip
	for _, v := range service.Spec.Ports {
		if v.Protocol == "" {
			v.Protocol = "TCP"
		}
	}
	body, _ = json.Marshal(service)
	err = s.store.Put(config.ServiceConfigPrefix+"/"+service.MetaData.Name, body)
	if err != nil {
		fmt.Println("[AddService] etcd put fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

func (s *Server) AddPod(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	pod := &object.Pod{}
	err = json.Unmarshal(body, pod)
	if err != nil {
		fmt.Println("[AddService] service unmarshal fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if pod.UID == "" {
		//这种情况下，可能etcd里边存的旧的有UID, 取出来填上，或者就是新的，需要分配
		//从etcd里取,看以前是否有过
		res, err2 := s.store.Get(config.PodConfigPREFIX + "/" + pod.Name)
		if err2 != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if len(res) != 0 {
			oldPod := &object.Pod{}
			err = json.Unmarshal(res[0].ValueBytes, oldPod)
			if err != nil {
				fmt.Println("[AddService] service unmarshal fail")
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			pod.UID = oldPod.UID
			if oldPod.Spec.NodeName != "" {
				pod.Spec.NodeName = oldPod.Spec.NodeName
			}
		} else {
			pod.UID = uuid.New().String()
		}
	}
	body, _ = json.Marshal(pod)

	err = s.store.Put(config.PodConfigPREFIX+"/"+pod.Name, body)
	if err != nil {
		fmt.Println("[AddService] etcd put fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
}
func (s *Server) AddDnsAndTrans(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	dnsAndTrans := &object.DnsAndTrans{}
	err = json.Unmarshal(body, dnsAndTrans)
	if err != nil {
		fmt.Println("[AddDnsAndTrans] Unmarshal fail " + err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if dnsAndTrans.Status.Phase == "" {
		//用户文件发来，此刻需要取出phase以及gateWayIp
		old, err3 := s.store.Get(config.DnsAndTransPrefix + "/" + dnsAndTrans.MetaData.Name)
		if err3 != nil {
			fmt.Println("[AddDnsAndTrans] add fail " + err3.Error())
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if len(old) != 0 {
			oldTrans := &object.DnsAndTrans{}
			json.Unmarshal(old[0].ValueBytes, oldTrans)
			if oldTrans.Status.Phase != object.Delete {
				dnsAndTrans.Status.Phase = oldTrans.Status.Phase
				dnsAndTrans.Spec.GateWayIp = oldTrans.Spec.GateWayIp
			}
		}
	}
	//获取所有的service进行, 从而回填ip
	res2, err2 := s.store.PrefixGet(config.ServicePrefix)
	if err2 != nil {
		fmt.Println("[AddDnsAndTrans] add fail " + err2.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var services []*object.Service
	for _, v := range res2 {
		s := &object.Service{}
		err = json.Unmarshal(v.ValueBytes, s)
		if err != nil {
			fmt.Println("[AddDnsAndTrans] add fail " + err.Error())
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		services = append(services, s)
	}
	for k, v := range dnsAndTrans.Spec.Paths {
		exist := false
		for _, s := range services {
			if v.Name == s.MetaData.Name {
				dnsAndTrans.Spec.Paths[k].Ip = s.Spec.ClusterIp
				exist = true
				break
			}
		}
		if exist {
			continue
		} else {
			//不存在对应的服务，应当报错
			fmt.Println("[AddDnsAndTrans] service not exist ")
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	body, _ = json.Marshal(dnsAndTrans)
	err = s.store.Put(config.DnsAndTransPrefix+"/"+dnsAndTrans.MetaData.Name, body)
	if err != nil {
		fmt.Println("[AddDnsAndTrans] etcd put fail")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	return

func (s *Server) getActivePods(ctx *gin.Context) {
	rsName := ctx.Query("rsName")
	if rsName == "" {
		fmt.Println("[getActivePods] rsName exist")
		ctx.Status(http.StatusBadRequest)
		return
	}

	uid := ctx.Query("uid")
	if uid == "" {
		fmt.Println("[getActivePods] uid not exist")
		ctx.Status(http.StatusBadRequest)
		return
	}

	var expect int
	var actual int

	listRes, err := s.store.PrefixGet("/registry/pod/default")
	if err != nil {
		fmt.Printf("[getActivePods] list fail\n")
		ctx.Status(http.StatusBadRequest)
		return
	}
	allPods, _ := makePods(listRes, rsName, uid)
	activePods := controller.FilterActivePods(allPods)
	actual = len(activePods)

	key := "/registry/rs/default/" + rsName
	raw, err := s.store.Get(key)
	if err != nil {
		fmt.Printf("[getActivePods] fail to get nodes\n")
	}
	if len(raw) == 0 {
		fmt.Printf("[getActivePods] list fail\n")
		ctx.Status(http.StatusBadRequest)
		return
	}
	result := &object.ReplicaSet{}

	err = json.Unmarshal(raw[0].ValueBytes, result)

	if err != nil {
		fmt.Printf("[getActivePods] unmarshal fail\n")
		ctx.Status(http.StatusBadRequest)
		return
	}

	expect = int(result.Spec.Replicas)

	ctx.JSON(200, gin.H{
		"expect": expect,
		"actual": actual,
	})
}

func makePods(raw []etcdstore.ListRes, name string, UID string) ([]*object.Pod, error) {
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

func ownBy(ownerReferences []object.OwnerReference, owner string, UID string) bool {
	for _, ref := range ownerReferences {
		if ref.Name == owner && ref.UID == UID {
			return true
		}
	}
	return false
}
