package queue

type RingQueue[TYPE any] struct {
	list     []TYPE
	capacity int
	tail     int
	size     int
}

func NewRingQueue[TYPE any](capacity int) *RingQueue[TYPE] {
	return &RingQueue[TYPE]{
		capacity: capacity,
		list:     make([]TYPE, capacity),
		tail:     0,
		size:     0,
	}
}

func (rq *RingQueue[TYPE]) Push(value TYPE) {
	rq.list[rq.tail] = value
	rq.tail = (rq.tail + 1) % rq.capacity
	rq.size++
	if rq.size >= rq.capacity {
		rq.size = rq.capacity
	}
}

func (rq *RingQueue[TYPE]) Len() int {
	return rq.size
}

func (rq *RingQueue[TYPE]) GetElements() []TYPE {
	return rq.list[:rq.size]
}
