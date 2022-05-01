package file

import (
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"reflect"
)

// UnmarshalPaths unmarshal a set of file
func UnmarshalPaths(v interface{}, paths []string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("bad interface")
	}

	notfound := true

	for _, p := range paths {
		if !fileExist(p) {
			continue
		}
		notfound = false
		if err := UnmarshalFile(v, p); err != nil {
			return err
		}
	}

	if notfound {
		return errors.New("no files found")
	}

	return nil
}

func UnmarshalFile(v interface{}, file string) error {
	if !fileExist(file) {
		fmt.Println("file not exist")
		return errors.New("file not exist")
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Read file error")
	}

	err = yaml.Unmarshal(buf, v)
	if err != nil {
		fmt.Printf("Unmarshal file error")
	}
	return nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
