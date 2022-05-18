package types

import "minik8s/object"

type PodOperation int

const (
	// SET is the current pod configuration.
	SET PodOperation = iota
	// ADD signifies pods that are new to this source.
	ADD
	// DELETE signifies pods that are gracefully deleted from this source.
	DELETE
	// UPDATE signifies pods have been updated in this source.
	UPDATE
)
const (
	// Filesource idenitified updates from a file.
	FileSource = "file"
	// HTTPSource identifies updates from querying a web page.
	HTTPSource = "http"
	// ApiserverSource identifies updates from Kubernetes API Server.
	ApiserverSource = "api"
	// AllSource identifies updates from all sources.
	AllSource = "*"
)

type PodUpdate struct {
	Pods   []*object.Pod
	Op     PodOperation
	Source string
}
