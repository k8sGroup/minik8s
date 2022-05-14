package object

type Autoscaler struct {
	Metadata ObjectMeta `json:"metadata" yaml:"metadata"`
	Spec     HPASpec    `json:"spec" yaml:"spec"`
}

type HPASpec struct {
	ScaleTargetRef HPARef  `json:"scaleTargetRef" yaml:"scaleTargetRef"`
	MinReplicas    int32   `json:"minReplicas" yaml:"minReplicas"`
	MaxReplicas    int32   `json:"maxReplicas" yaml:"maxReplicas"`
	Metrics        Metrics `json:"metrics" yaml:"metrics"`
}

type HPARef struct {
	// optional value : app.Deployment, app.Replicaset
	Kind string `json:"kind" yaml:"kind"`
	Name string `json:"name" yaml:"name"`
}

type Metrics struct {
	Cpu    *CpuMetric    `json:"cpu" yaml:"cpu"`       // this field could be nil !
	Memory *MemoryMetric `json:"memory" json:"memory"` // this field could be nil !
}

type CpuMetric struct {
	AverageUtilization int32 `json:"averageUtilization" yaml:"averageUtilization"`
}

type MemoryMetric struct {
	// example : 100mb 100MB 1gb 1GB
	AverageUsage string `json:"averageUsage" yaml:"averageUsage"`
}
