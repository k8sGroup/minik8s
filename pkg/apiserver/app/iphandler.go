package app

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcdstore/nodeConfigStore"
	"minik8s/pkg/klog"
	"net/http"
	"time"
)

func (s *Server) addNode(ctx *gin.Context) {
	physicalIp := ctx.Param(config.ParamResourceName)
	klog.Infof("%s, physicalIp received is %s", time.Now().Format("2006-01-02 15:04:05"), physicalIp)
	node, err := nodeConfigStore.AddNewNode(physicalIp)
	if err != nil {
		klog.Errorf("%s, %s", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	key := config.NODE_PREFIX + "/" + physicalIp
	raw, _ := json.Marshal(node)
	err = s.store.Put(key, raw)
	if err != nil {
		klog.Errorf("%s, %s", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *Server) deleteNode(ctx *gin.Context) {
	physicalIp := ctx.Param(config.ParamResourceName)
	nodeConfigStore.DeleteNode(physicalIp)
	key := config.NODE_PREFIX + "/" + physicalIp
	err := s.store.Del(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}
