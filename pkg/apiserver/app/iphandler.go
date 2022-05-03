package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"minik8s/pkg/apiserver/config"
	"net/http"
)

func (s *Server) addNode(ctx *gin.Context) {
	clusterIp := ctx.Param(config.ParamClusterIp)
	pair, err := s.netConfigStore.AddNewNode(clusterIp)
	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	key := "/node/register/" + clusterIp
	raw, _ := json.Marshal(pair)
	err = s.store.Put(key, raw)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}
