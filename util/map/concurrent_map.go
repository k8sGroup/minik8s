package concurrent_map

import (
	"sync"
)

type ConcurrentMap struct {
	m   map[interface{}]interface{}
	mtx sync.RWMutex
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		m: make(map[interface{}]interface{})}
}

func (cp *ConcurrentMap) Get(key interface{}) interface{} {
	cp.mtx.RLock()
	defer cp.mtx.RUnlock()
	return cp.m[key]
}

func (cp *ConcurrentMap) Put(key interface{}, elem interface{}) {
	cp.mtx.Lock()
	defer cp.mtx.Unlock()
	cp.m[key] = elem
}

func (cp *ConcurrentMap) Contains(key interface{}) bool {
	cp.mtx.RLock()
	defer cp.mtx.RUnlock()
	_, ok := cp.m[key]
	return ok
}
