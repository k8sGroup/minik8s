package app

import (
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

type Store struct {
	client *etcd.Client
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

func (s *Store) Get(key string) ([]byte, error) {
	kv := etcd.NewKV(s.client)
	response, err := kv.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 {
		return []byte{}, nil
	} else {
		return response.Kvs[0].Value, nil
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

func (s *Store) PrefixGet(key string) ([][]byte, error) {
	kv := etcd.NewKV(s.client)
	response, err := kv.Get(context.TODO(), key, etcd.WithPrefix())
	if err != nil {
		return nil, err
	}
	var ret [][]byte
	for _, kv := range response.Kvs {
		ret = append(ret, kv.Value)
	}
	return ret, nil
}
