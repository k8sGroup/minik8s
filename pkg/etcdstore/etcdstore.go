package etcdstore

import (
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"minik8s/pkg/klog"
	"time"
)

type Store struct {
	client *etcd.Client
}
type WatchResType int

const (
	PUT    WatchResType = 0
	DELETE WatchResType = 1
)

type WatchRes struct {
	ResType         WatchResType
	ResourceVersion int64
	ValueBytes      []byte
}

type ListRes struct {
	ResourceVersion int64
	ValueBytes      []byte
}

func NewEtcdStore(endpoints []string, timeout time.Duration) (*Store, error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = cli.Status(timeoutContext, endpoints[0])
	if err != nil {
		return nil, err
	}
	return &Store{client: cli}, nil
}

func (s *Store) Get(key string) ([]ListRes, error) {
	kv := etcd.NewKV(s.client)
	response, err := kv.Get(context.TODO(), key)
	if err != nil {
		return []ListRes{}, err
	}
	if len(response.Kvs) == 0 {
		return []ListRes{}, nil
	} else {
		listRes := ListRes{ResourceVersion: response.Kvs[0].ModRevision, ValueBytes: response.Kvs[0].Value}
		return []ListRes{listRes}, nil
	}
}

func (s *Store) Put(key string, val []byte) error {
	kv := etcd.NewKV(s.client)
	_, err := kv.Put(context.TODO(), key, string(val))
	return err
}

func (s *Store) Del(key string) error {
	kv := etcd.NewKV(s.client)
	_, err := kv.Delete(context.TODO(), key)
	return err
}

func (s *Store) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	watch := func(c chan<- WatchRes) {
		watchChan := s.client.Watch(ctx, key)
		for watchResponse := range watchChan {
			for _, event := range watchResponse.Events {
				var res WatchRes
				switch event.Type {
				case etcd.EventTypePut:
					res.ResType = PUT
					res.ResourceVersion = event.Kv.ModRevision
					res.ValueBytes = event.Kv.Value
					break
				case etcd.EventTypeDelete:
					res.ResType = DELETE
					res.ResourceVersion = event.Kv.ModRevision
					break
				}
				c <- res
			}
		}
		klog.Infof("Closing watching channel for key %s\n", key)
		close(c)
	}
	go watch(watchResChan)

	return cancel, watchResChan
}

func (s *Store) PrefixWatch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	watch := func(c chan<- WatchRes) {
		watchChan := s.client.Watch(ctx, key, etcd.WithPrefix())
		for watchResponse := range watchChan {
			for _, event := range watchResponse.Events {
				var res WatchRes
				switch event.Type {
				case etcd.EventTypePut:
					res.ResType = PUT
					res.ResourceVersion = event.Kv.ModRevision
					res.ValueBytes = event.Kv.Value
					break
				case etcd.EventTypeDelete:
					res.ResType = DELETE
					res.ResourceVersion = event.Kv.ModRevision
					break
				}
				c <- res
			}
		}
		klog.Infof("Closing prefix watching channel for key %s\n", key)
		close(c)
	}
	go watch(watchResChan)
	return cancel, watchResChan
}

func (s *Store) PrefixGet(key string) ([]ListRes, error) {
	kv := etcd.NewKV(s.client)
	response, err := kv.Get(context.TODO(), key, etcd.WithPrefix())
	if err != nil {
		return []ListRes{}, err
	}
	var ret []ListRes
	for _, kv := range response.Kvs {
		res := ListRes{
			ResourceVersion: kv.ModRevision,
			ValueBytes:      kv.Value,
		}
		ret = append(ret, res)
	}
	return ret, nil
}
