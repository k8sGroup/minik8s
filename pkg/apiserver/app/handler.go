package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"net/http"
)

//
//// do not delete pod in etcd directly, just modify the status
//// TODO: real deletion by kubelet !
//func (s *Server) deletePod(ctx *gin.Context) {
//	name := ctx.Param(config.ParamResourceName)
//	key := "/registry/pod/default/" + name
//	resList, err := s.store.Get(key)
//	if err != nil || len(resList) == 0 {
//		fmt.Printf("[deletePod] pod not exist:%s\n", name)
//		ctx.AbortWithStatus(http.StatusBadRequest)
//		return
//	}
//	pod := object.Pod{}
//	err = json.Unmarshal(resList[0].ValueBytes, &pod)
//	if err != nil {
//		fmt.Printf("[deletePod] pod unmarshal fail\n")
//		ctx.AbortWithStatus(http.StatusBadRequest)
//		return
//	}
//	// update pod phases to failed
//	pod.Status.Phase = object.PodFailed
//	raw, _ := json.Marshal(pod)
//	err = s.store.Put(key, raw)
//	if err != nil {
//		ctx.AbortWithStatus(http.StatusBadRequest)
//	}
//	ctx.Status(http.StatusOK)
//}

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

	// set spec replicas to zero
	rs.Spec.Replicas = 0
	raw, _ := json.Marshal(rs)
	err = s.store.Put(key, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}
