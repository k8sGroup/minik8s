package lister

import (
	"minik8s/object"
	"minik8s/pkg/labels"
)

type ReplicaSetLister interface {
	ReplicaSets(namespace string) ReplicaSetNamespaceLister
}

type ReplicaSetNamespaceLister interface {
	List(selector labels.Selector) (ret []*object.ReplicaSet, err error)
	Get(name string) (*object.ReplicaSet, error)
}
