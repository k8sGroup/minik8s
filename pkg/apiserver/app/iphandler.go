package app

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/klog"
	"net/http"
	"time"
)

func (s *Server) addNode(ctx *gin.Context) {
	physicalIp := ctx.Param(config.ParamResourceName)
	klog.Infof("%s, physicalIp received is %s", time.Now().Format("2006-01-02 15:04:05"), physicalIp)
	pair, err := s.netConfigStore.AddNewNode(physicalIp)
	if err != nil {
		klog.Errorf("%s, %s", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	key := config.NODE_PREFIX + "/" + physicalIp
	raw, _ := json.Marshal(pair)
	err = s.store.Put(key, raw)
	if err != nil {
		klog.Errorf("%s, %s", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}
