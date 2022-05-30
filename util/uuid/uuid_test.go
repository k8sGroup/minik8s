package uuid

import (
	"fmt"
	"testing"
)

func TestNewUUID(t *testing.T) {
	for i := 0; i < 5; i++ {
		uuid := NewUUID(5)
		fmt.Println(uuid)
	}

}
