package tools

import (
	"os"
	"path"
)

func RemoveAll(localPath string) error {
	info, err := os.Stat(localPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	if !info.IsDir() {
		return os.Remove(localPath)
	} else {
		fileInfos, err := os.ReadDir(localPath)
		if err != nil {
			return err
		}
		for _, info := range fileInfos {
			subPath := path.Join(localPath, info.Name())
			err = RemoveAll(subPath)
		}
		err = os.Remove(localPath)
		if err != nil {
			return err
		}
	}
	return nil
}
