package app

import (
	"bytes"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"minik8s/pkg/apiserver/config"
	"net/http"
)

type Server struct {
	engine      *gin.Engine
	port        int
	resourceSet mapset.Set[string]
	store       *Store
}

func NewServer(c *config.Config) (*Server, error) {
	engine := gin.Default()
	store, err := NewEtcdStore(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		return nil, err
	}
	s := &Server{
		engine:      engine,
		port:        c.HttpPort,
		resourceSet: mapset.NewSet[string](c.ValidResources...),
		store:       store,
	}

	{
		engine.GET(config.Path, s.validate, s.get)
		engine.DELETE(config.Path, s.validate, s.del)
		engine.PUT(config.Path, s.validate, s.put)
	}
	{
		engine.GET(config.PrefixPath, s.validate, s.prefixGet)
	}

	return s, nil
}

func (s *Server) Run() error {
	// start web api
	err := s.engine.Run(fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) validate(c *gin.Context) {
	resource := c.Param(config.ParamResource)
	if !s.resourceSet.Contains(resource) {
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

func (s *Server) get(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	value, err := s.store.Get(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Data(http.StatusOK, "application/json", value)
}

func (s *Server) put(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	err = s.store.Put(key, body)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *Server) del(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	err := s.store.Del(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	ctx.Status(http.StatusOK)
}

func (s *Server) prefixGet(ctx *gin.Context) {
	prefixKey := ctx.Request.URL.Path
	vals, err := s.store.PrefixGet(prefixKey)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	merged := bytes.Join(vals, []byte{','})
	var data []byte
	data = append([]byte{'['}, merged...)
	data = append(data, ']')
	ctx.Data(http.StatusOK, "application/json", data)
}
