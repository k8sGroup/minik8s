package tools

import (
	"bytes"
	"io"
	"os"
	"path"
)

func Bytes2File(raw []byte, name string, dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	file, err := os.OpenFile(path.Join(dirPath, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = io.Copy(file, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	return nil
}
