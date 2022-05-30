package object

const (
	MetricCPU     string = "cpu"
	MetricMemory  string = "memory"
	MetricMax     string = "max"
	MetricAverage string = "average"
)

type Autoscaler struct {
	Metadata ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     HPASpec    `json:"spec" yaml:"spec"`
}

type HPASpec struct {
	ScaleTargetRef HPARef   `json:"scaleTargetRef" yaml:"scaleTargetRef"`
	MinReplicas    int32    `json:"minReplicas" yaml:"minReplicas"`
	MaxReplicas    int32    `json:"maxReplicas" yaml:"maxReplicas"`
	ScaleInterval  int32    `json:"scaleInterval" yaml:"scaleInterval"`
	Metrics        []Metric `json:"metrics" yaml:"metrics"`
}

type HPARef struct {
	// optional value : app.Deployment, app.Replicaset
	Kind string `json:"kind" yaml:"kind"`
	Name string `json:"name" yaml:"name"`
}

type Metric struct {
	Name       string `json:"name" yaml:"name"`             // MetricCPU or MetricMemory
	Strategy   string `json:"strategy" yaml:"strategy"`     // MetricMax or MetricAverage
	Percentage int32  `json:"percentage" yaml:"percentage"` // percentage/100
}
