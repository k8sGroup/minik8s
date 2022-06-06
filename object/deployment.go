package object

type Deployment struct {
	Metadata ObjectMeta     `json:"metadata" yaml:"metadata"`
	Spec     DeploymentSpec `json:"spec" yaml:"spec"`
}

type DeploymentSpec struct {
	Replicas int32               `json:"replicas" yaml:"replicas"`
	Strategy *DeploymentStrategy `json:"strategy" yaml:"strategy"`
	Template PodTemplate         `json:"template" yaml:"template"`
}

type DeploymentStrategy struct {
	Type          string                 `json:"type" yaml:"type"`
	RollingUpdate *StrategyRollingUpdate `json:"rollingUpdate" yaml:"rollingUpdate"`
}

type StrategyRollingUpdate struct {
	MaxSurge       *int32 `json:"maxSurge" yaml:"maxSurge"`
	MaxUnavailable *int32 `json:"maxUnavailable" yaml:"maxUnavailable"`
}

func (d *Deployment) Complete() {
	d.Spec.Complete()
}

func (ds *DeploymentSpec) Complete() {
	if ds.Strategy == nil {
		ds.Strategy = &DeploymentStrategy{Type: "rollingUpdate"}
	}
	ds.Strategy.Complete()
}

func (ds *DeploymentStrategy) Complete() {
	if ds.RollingUpdate == nil {
		ds.RollingUpdate = &StrategyRollingUpdate{}
	}
	ds.RollingUpdate.Complete()
}

func (sru *StrategyRollingUpdate) Complete() {
	if sru.MaxSurge == nil {
		sru.MaxSurge = new(int32)
		*sru.MaxSurge = 1
	}
	if sru.MaxUnavailable == nil {
		sru.MaxUnavailable = new(int32)
		*sru.MaxUnavailable = 1
	}
}
