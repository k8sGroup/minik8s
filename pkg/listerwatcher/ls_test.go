package listerwatcher

import (
	"fmt"
	"minik8s/pkg/etcdstore"
	"testing"
	"time"
)

func TestLs(t *testing.T) {
	ls, err := NewListerWatcher(DefaultConfig())
	if err != nil {
		fmt.Println(err)
		return
	}
	cancelFunc, err := ls.WatchNonBlocking("/registry/test/default", func(res etcdstore.WatchRes) {
		fmt.Println(string(res.ValueBytes))
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(40 * time.Second)
	cancelFunc()
	time.Sleep(3 * time.Second)
}
