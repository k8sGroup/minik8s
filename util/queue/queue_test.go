package queue

import (
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	que := ConcurrentQueue{}
	strs := [...]string{"aaa", "bbb", "ccc"}
	for _, str := range strs {
		que.Enqueue(str)
	}

	for i, _ := range strs {
		if strs[i] != que.Front() {
			assert.Error(t, errors.New("queue wrong value"), "")
		}
		que.Dequeue()
	}

	if que.Empty() != true {
		assert.Error(t, errors.New("queue not empty"), "")
	}
}
