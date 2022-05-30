package queue

import (
	"fmt"
	"testing"
)

func TestRingQueue(t *testing.T) {
	ringQ := NewRingQueue[float64](5)
	ringQ.Push(22.8)
	ringQ.Push(66.1)
	vals := ringQ.GetElements()
	for _, i := range vals {
		fmt.Println(i)
	}
	fmt.Println("")
	ringQ.Push(90.7)
	ringQ.Push(11.1)
	ringQ.Push(49.2)
	ringQ.Push(62.1)
	vals = ringQ.GetElements()
	for _, i := range vals {
		fmt.Println(i)
	}
	fmt.Println("")
}
