package klog

import "testing"

func TestInfof(t *testing.T) {
	Infof("aaa %d", 1)
}
