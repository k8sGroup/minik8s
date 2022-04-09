package client

import (
	"context"
	"encoding/json"
	"minik8s/k8s.io/client/rest"
	"minik8s/object"
)

// ReplicaSetInterface has methods to work with ReplicaSet resources.
type ReplicaSetInterface interface {
	Create(ctx context.Context, replicaSet *object.ReplicaSet) (*object.ReplicaSet, error)
}

// replicaSets implements ReplicaSetInterface
type replicaSets struct {
	client *rest.RESTClient
	ns     string
}

// newReplicaSets returns a ReplicaSets
func newReplicaSets(c *rest.RESTClient, namespace string) *replicaSets {
	return &replicaSets{
		client: c,
		ns:     namespace,
	}
}

// Create takes the representation of a replicaSet and creates it.  Returns the server's representation of the replicaSet, and an error, if there is any.
func (c *replicaSets) Create(ctx context.Context, replicaSet *object.ReplicaSet) (result *object.ReplicaSet, err error) {
	result = &object.ReplicaSet{}
	res := c.client.Post().
		Namespace(c.ns).
		Body(replicaSet).
		Do(ctx)
	body := res.GetBody()
	err = json.Unmarshal(body, result)
	return
}
