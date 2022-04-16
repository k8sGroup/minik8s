package pod

import (
	"github.com/satori/go.uuid"
	"minik8s/cmd/kubelet/app/module"
	"os"
)

const emptyDir = "emptyDir"
const hostPath = "hostPath"

type Pod struct {
	Name string
	//LabelMap     map[string]string
	ContainerIds []string
	TmpDirMap    map[string]string
	HostDirMap   map[string]string
}

func (p Pod) AddVolumes(volumes []module.Volume) error {
	p.TmpDirMap = make(map[string]string)
	p.HostDirMap = make(map[string]string)
	for _, value := range volumes {
		if value.Type == emptyDir {
			//临时目录，随机生成
			u := uuid.NewV4()
			path := "./tmp/" + u.String()
			os.MkdirAll(path, os.ModePerm)
			p.TmpDirMap[value.Name] = path
		} else {
			//指定了实际目录
			_, err := os.Stat(value.Path)
			if err != nil {
				os.MkdirAll(value.Path, os.ModePerm)
			}
			p.HostDirMap[value.Name] = value.Path
		}
	}
	return nil
}
