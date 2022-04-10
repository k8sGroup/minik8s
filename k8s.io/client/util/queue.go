package workqueue

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
}
