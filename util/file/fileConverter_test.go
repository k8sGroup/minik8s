package file

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"minik8s/object"
	"testing"
)

var (
	test_path = "../../test"
)

func TestUnmarshalFile(t *testing.T) {
	path := test_path + "/pod.yaml"
	pod := &object.Pod{}
	err := UnmarshalFile(pod, path)
	assert.Equal(t, err, nil)
	fmt.Printf("pod:%+v", pod)
}
