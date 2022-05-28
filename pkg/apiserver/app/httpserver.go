package app

import (
	"context"
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"io/ioutil"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/messaging"
	"net/http"
	"strconv"
	"sync"
)

type watchOpt struct {
	key        string
	withPrefix bool
	ticket     uint64
}

type Ticket struct {
	T uint64
}

type Server struct {
	engine       *gin.Engine
	port         int
	resourceSet  mapset.Set[string]
	store        *etcdstore.Store
	publisher    *messaging.Publisher
	watcherMap   map[string]*watcher
	watcherMtx   sync.Mutex // watcherMtx 保护watcherCount
	watcherChan  chan watchOpt
	ticketSeller *atomic.Uint64
}

type watcher struct {
	set    mapset.Set[uint64]
	cancel context.CancelFunc
}

func NewServer(c *config.ServerConfig) (*Server, error) {
	engine := gin.Default()
	store, err := etcdstore.NewEtcdStore(c.EtcdEndpoints, c.EtcdTimeout)
	if err != nil {
		return nil, err
	}
	publisher, err := messaging.NewPublisher(c.QueueConfig)
	if err != nil {
		return nil, err
	}
	watcherChan := make(chan watchOpt)
	//kubeNetSupport, err2 := kubeNetSupport.NewKubeNetSupport(listerwatcher.DefaultConfig(), client.DefaultClientConfig())
	//if err2 != nil {
	//	return nil, err2
	//}
	s := &Server{
		engine:       engine,
		port:         c.HttpPort,
		resourceSet:  mapset.NewSet[string](c.ValidResources...),
		store:        store,
		publisher:    publisher,
		watcherMap:   map[string]*watcher{},
		watcherChan:  watcherChan,
		ticketSeller: atomic.NewUint64(0),
		//kubeNetSupport: kubeNetSupport,
	}

	{
		engine.GET(config.Path, s.validate, s.get)
		engine.DELETE(config.Path, s.validate, s.del)
		engine.PUT(config.Path, s.validate, s.put)
		engine.POST(config.Path, s.validate, s.watch)
	}
	{
		engine.GET(config.PrefixPath, s.validate, s.prefixGet)
		engine.POST(config.PrefixPath, s.validate, s.prefixWatch)
	}

	{
		engine.DELETE(config.RS, s.deleteRS)

	}
	{
		engine.PUT(config.PodCONFIG, s.AddPod)
		engine.DELETE(config.PodCONFIG, s.deletePod)
	}
	{
		engine.PUT(config.NODE, s.addNode)
		engine.GET(config.NODE, s.get)
		engine.POST(config.NODE, s.watch)
		engine.DELETE(config.NODE, s.deleteNode)
	}
	{
		engine.GET(config.NODE_PREFIX, s.prefixGet)
		engine.POST(config.NODE_PREFIX, s.prefixWatch)
	}
	{
		// user operation
		engine.PUT(config.UserPodPath, s.userAddPod)
		engine.PUT(config.UserRSPath, s.userAddRS)
	}
	{
		engine.PUT(config.ServiceConfig, s.AddService)
		engine.DELETE(config.ServiceConfig, s.deleteService)
	}
	{
		engine.GET(config.RS_POD, s.getActivePods)
	}

	go s.daemon(watcherChan)

	return s, nil
}

func (s *Server) Run() error {
	// start web api
	err := s.engine.Run(fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	//err = s.kubeNetSupport.StartKubeNetSupport()
	return err
}

func (s *Server) validate(c *gin.Context) {
	resource := c.Param(config.ParamResource)
	if !s.resourceSet.Contains(resource) {
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

func (s *Server) get(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	listRes, err := s.store.Get(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	data, err := json.Marshal(listRes)
	ctx.Data(http.StatusOK, "application/json", data)
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
	listResList, err := s.store.PrefixGet(prefixKey)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	data, err := json.Marshal(listResList)
	ctx.Data(http.StatusOK, "application/json", data)
}

/*
watch
根据key进行watch，通知server注册一个rabbitmq的消息队列，返回值为200+T。

client收到返回值之后检查是否是200，如果是200
新建一个 Subscriber 或者使用原有的 Subscriber 进行 subscribe 操作。

订阅一个交换机，交换机的名字为watch的路径

注意，在watch结束时必须发送一个POST请求，表示watch已经结束。
否则会造成服务器长时间watch导致资源泄漏！

路径为之前prefixWatch的路径，附带有form-data参数
参数名为ticket，参数值为之前收到的ticket

*/
func (s *Server) watch(ctx *gin.Context) {
	fmt.Println("watch match...")
	key := ctx.Request.URL.Path
	ticketStr, status := ctx.GetPostForm("ticket")
	if !status {
		t := Ticket{}
		t.T = s.ticketSeller.Add(1)
		data, _ := json.Marshal(t)
		s.watcherChan <- watchOpt{key: key, withPrefix: false, ticket: t.T}
		ctx.Data(http.StatusOK, "application/json", data)
	} else {
		s.watcherMtx.Lock()
		ticket, err := strconv.ParseUint(ticketStr, 10, 64)
		if err != nil {
			klog.Infof("%s\n", err.Error())
			ctx.AbortWithStatus(http.StatusBadRequest)
		} else {
			if s.watcherMap[key] != nil {
				s.watcherMap[key].set.Remove(ticket)
				if s.watcherMap[key].set.Equal(mapset.NewSet[uint64]()) {
					s.watcherMap[key].cancel()
					s.watcherMap[key] = nil
					klog.Infof("Cancel the watcher of key %s\n", key)
				}
			}
			ctx.Status(http.StatusOK)
		}
		s.watcherMtx.Unlock()
	}
}

/*
prefixWatch
根据前缀进行watch，通知server注册一个rabbitmq的消息队列，返回值为200+T。

client收到返回值之后检查是否是200，如果是200
新建一个 Subscriber 或者使用原有的 Subscriber 进行 subscribe 操作。

订阅一个交换机，交换机的名字为prefixWatch的路径

注意，在watch结束时必须发送一个POST请求，表示watch已经结束。
否则会造成服务器长时间watch导致资源泄漏！

路径为之前prefixWatch的路径，附带有form-data参数
参数名为ticket，参数值为之前收到的ticket

*/
func (s *Server) prefixWatch(ctx *gin.Context) {
	key := ctx.Request.URL.Path
	ticketStr, status := ctx.GetPostForm("ticket")
	if !status {
		t := Ticket{}
		t.T = s.ticketSeller.Add(1)
		data, _ := json.Marshal(t)
		s.watcherChan <- watchOpt{key: key, withPrefix: true, ticket: t.T}
		ctx.Data(http.StatusOK, "application/json", data)
	} else {
		s.watcherMtx.Lock()
		ticket, err := strconv.ParseUint(ticketStr, 10, 64)
		if err != nil {
			klog.Infof("%s\n", err.Error())
			ctx.AbortWithStatus(http.StatusBadRequest)
		} else {
			if s.watcherMap[key] != nil {
				s.watcherMap[key].set.Remove(ticket)
				if s.watcherMap[key].set.Equal(mapset.NewSet[uint64]()) {
					s.watcherMap[key].cancel()
					s.watcherMap[key] = nil
					klog.Infof("Cancel the prefix watcher of key %s\n", key)
				}
			}
			ctx.Status(http.StatusOK)
		}
		s.watcherMtx.Unlock()
	}
}

func (s *Server) daemon(listening <-chan watchOpt) {
	for {
		select {
		case opt := <-listening:
			key := opt.key
			ticket := opt.ticket
			withPrefix := opt.withPrefix
			{
				// critical section
				s.watcherMtx.Lock()
				if w := s.watcherMap[key]; w == nil {
					var cancel context.CancelFunc
					var resChan <-chan etcdstore.WatchRes
					w := &watcher{set: mapset.NewSet[uint64]()}
					if withPrefix {
						cancel, resChan = s.store.PrefixWatch(key)
					} else {
						cancel, resChan = s.store.Watch(key)
					}
					w.cancel = cancel
					s.watcherMap[key] = w
					s.watcherMap[key].set.Add(ticket)

					go func(resChan <-chan etcdstore.WatchRes) {
						for res := range resChan {
							data, err := json.Marshal(res)
							if err != nil {
								klog.Errorf("%v\n", err)
							}
							err = s.publisher.Publish(key, data, "application/json")
							if err != nil {
								klog.Errorf("%v\n", err)
							}
						}
						klog.Infof("Res Chan closed for key %s\n", key)
					}(resChan)
				} else {
					s.watcherMap[key].set.Add(ticket)
				}
				s.watcherMtx.Unlock()
			}
		}
	}
}
