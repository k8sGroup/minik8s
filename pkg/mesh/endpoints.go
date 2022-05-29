package mesh

import (
	"github.com/pkg/errors"
	"math/rand"
	"sync"
	"time"
)

type EndPoint struct {
	PodIP  string
	Weight int
}

type Dispatcher struct {
	m   map[string][]EndPoint
	mtx sync.RWMutex
}

func NewDispatcher() *Dispatcher {
	rand.Seed(time.Now().Unix())
	return &Dispatcher{}
}

func (d *Dispatcher) UpsertEndpoints(clusterIP string, podIP string, weight int) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	endpoints, ok := d.m[clusterIP]

	if !ok {
		d.m[clusterIP] = []EndPoint{{podIP, weight}}
	} else {
		for _, ep := range endpoints {
			if ep.PodIP == podIP {
				ep.Weight = weight
				return
			}
		}
		d.m[clusterIP] = append(d.m[clusterIP], EndPoint{podIP, weight})
	}
}

func (d *Dispatcher) GetEndPoint(clusterIP string) (podIP *string, err error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	endpoints, ok := d.m[clusterIP]
	if !ok || len(endpoints) == 0 {
		return nil, errors.New("no endpoints")
	}

	var sum int
	for _, ep := range endpoints {
		sum += ep.Weight
	}

	num := rand.Intn(sum) + 1
	sum = 0
	for _, ep := range endpoints {
		sum += ep.Weight
		if sum >= num {
			return &ep.PodIP, nil
		}
	}

	return nil, errors.New("no endpoints chosen")
}
