package concurrent_map

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestConcurrentMap(t *testing.T) {
	cp := NewConcurrentMap()
	cp.Put("aaa", 1)
	assert.Equal(t, 1, cp.Get("aaa"))
}
