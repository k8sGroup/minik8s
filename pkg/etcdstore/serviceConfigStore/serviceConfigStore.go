package serviceConfigStore

import (
	"fmt"
	"sync"
)

//主要用于分配clusterIp,当yaml为空时
const (
	BaseClusterIp string = "10.10"
)

type ServiceConfigStore struct {
	//service Name到ClusterIp的映射
	Name2ClusterIp map[string]string
	BasicClusterIp string
	FourthField    int
	ThirdField     int
}

var instance *ServiceConfigStore
var lock sync.Locker

func newServiceConfigStore() *ServiceConfigStore {
	res := &ServiceConfigStore{}
	res.Name2ClusterIp = make(map[string]string)
	res.BasicClusterIp = BaseClusterIp
	res.FourthField = 0
	res.ThirdField = 0
	return res
}
func getServiceConfigStore() *ServiceConfigStore {
	if instance == nil {
		instance = newServiceConfigStore()
		return instance
	} else {
		return instance
	}
}
func isExist(m map[string]string, val string) bool {
	for _, v := range m {
		if v == val {
			return true
		}
	}
	return false
}
func (store *ServiceConfigStore) allocClusterIp() string {
	for {
		clusterIp := store.BasicClusterIp + "." + fmt.Sprintf("%d", store.ThirdField) + "." + fmt.Sprintf("%d", store.FourthField)
		store.FourthField++
		if store.FourthField == 256 {
			store.ThirdField++
			store.FourthField = 0
		}
		if isExist(store.Name2ClusterIp, clusterIp) {
			continue
		} else {
			return clusterIp
		}
	}
}

//判断是否合法以及分配ClusterIp

func JudgeAndAllocClusterIp(name string, clusterIp string) (bool, string) {
	lock.Lock()
	defer lock.Unlock()
	store := getServiceConfigStore()
	if clusterIp == "" {
		val, ok := store.Name2ClusterIp[name]
		if !ok {
			//需要分配
			ip := store.allocClusterIp()
			store.Name2ClusterIp[name] = ip
			return true, ip
		} else {
			return true, val
		}
	} else {
		//判断是否合规即可
		val, ok := store.Name2ClusterIp[name]
		if !ok {
			//map中不存在，需要判断clusterIp是否重复
			if isExist(store.Name2ClusterIp, clusterIp) {
				return false, clusterIp
			} else {
				store.Name2ClusterIp[name] = clusterIp
				return true, clusterIp
			}
		} else {
			//map中存在，需要更新, 同时也需要判断有无重复
			if val == clusterIp {
				return true, clusterIp
			} else {
				if isExist(store.Name2ClusterIp, clusterIp) {
					return false, clusterIp
				} else {
					store.Name2ClusterIp[name] = clusterIp
					return true, clusterIp
				}
			}
		}
	}
}
