package module

type Labels struct {
	App string `yaml:"app"`
}
type MetaData struct {
	Name string `yaml:"name"`
}
type Port struct {
	ContainerPort string `yaml:"containerPort"`
}
type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}
type Container struct {
	Name            string        `yaml:"name"`
	Image           string        `yaml:"image"`
	ImagePullPolicy string        `yaml:"imagePullPolicy"`
	Command         []string      `yaml:"command"`
	VolumeMounts    []VolumeMount `yaml:"volumeMounts"`
	Limits          Limit         `yaml:"limits"`
	Ports           []Port        `yaml:"ports"`
}
type Limit struct {
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}
type Volume struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}
type Spec struct {
	Volumes    []Volume    `yaml:"volumes"`
	Containers []Container `yaml:"containers"`
}
type Config struct {
	Kind     string   `yaml:"kind"`
	MetaData MetaData `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
}
