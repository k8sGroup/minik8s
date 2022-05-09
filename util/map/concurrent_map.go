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

type ConcurrentMapTrait[KEY, VALUE] struct {
	innerMap map[KEY]VALUE
	mtx      sync.RWMutex
}

func NewConcurrentMapTrait[KEY, VALUE]() *ConcurrentMapTrait[KEY, VALUE] {
	return &ConcurrentMapTrait[KEY, VALUE]{
		innerMap: make(map[KEY]VALUE),
	}
}

func (c *ConcurrentMapTrait[KEY, VALUE]) Get(key KEY) (VALUE, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	val, ok := c.innerMap[key]
	return val, ok
}

func (c *ConcurrentMapTrait[KEY, VALUE]) Put(key KEY, val VALUE) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.innerMap[key] = val
}

func (c *ConcurrentMapTrait[KEY, VALUE]) Del(key KEY) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.innerMap, key)
}

func (c *ConcurrentMapTrait[KEY, VALUE]) Contains(key KEY) bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	_, ok := c.innerMap[key]
	return ok
}

func (c *ConcurrentMapTrait[KEY, VALUE]) ReplaceAll(newMap map[KEY]VALUE) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.innerMap = newMap
}

func (c *ConcurrentMapTrait[KEY, VALUE]) SnapShot() map[KEY]VALUE {
	c.mtx.RUnlock()
	defer c.mtx.RUnlock()
	m2 := c.innerMap
	return m2
}
