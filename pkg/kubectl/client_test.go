package kubectl

import (
	"encoding/json"
	"fmt"
	"testing"
)

type testResource struct {
	Name string
	ID   int
}

func TestPut(t *testing.T) {
	r := testResource{Name: "pod-a", ID: 199}
	err := Put("http://localhost:8080/registry/test/default/pod-a", r)
	if err != nil {
		t.Errorf("%v\n", err)
	}
}

func TestGet(t *testing.T) {
	var r testResource
	resList, err := Get("http://localhost:8080/registry/test/default/pod-a")
	if err != nil {
		t.Errorf("%v\n", err)
	}
	err = json.Unmarshal(resList[0].ValueBytes, &r)
	if err != nil {
		t.Errorf("%v\n", err)
	}
	fmt.Printf("%v\n", resList)
	fmt.Println(r)
}
