package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
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
	key := "/registry/rs/default/" + name
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
		err = s.store.Del(key)
		ctx.Status(http.StatusOK)
		return
	}

	// set spec replicas to zero
	rs.Spec.Replicas = 0
	raw, _ := json.Marshal(rs)
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
		fmt.Printf("[deletePod] pod unmarshal fail\n")
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
	uid := uuid.New().String()
	service.MetaData.UID = uid
	if service.Spec.Type == "" {
		service.Spec.Type = object.ClusterIp
	}
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
