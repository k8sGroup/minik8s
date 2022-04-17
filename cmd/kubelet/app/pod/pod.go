package pod

import (
	"github.com/pquerna/ffjson/ffjson"
	"github.com/satori/go.uuid"
	"minik8s/cmd/kubelet/app/module"
	"os"
	"path"
	"runtime"
)

const emptyDir = "emptyDir"
const hostPath = "hostPath"

//pod的四种状态
const POD_PENDING_STATUS = "Pending"
const POD_FAILED_STATUS = "Failed"
const POD_RUNNING_STATUS = "Running"
const POD_SUCCEED_STATUS = "Succeed"

type Pod struct {
	Name string
	//LabelMap     map[string]string
	Uid string
	//create time
	Ctime        string
	ContainerIds []string
	TmpDirMap    map[string]string
	HostDirMap   map[string]string
	HostFileMap  map[string]string
	Status       string
	//如果有错的错误信息
	Err error
}

func NewPodfromConfig(config module.Config) *Pod {
	newPod := &Pod{}
	newPod.Name = config.MetaData.Name
	newPod.Uid = uuid.NewV4().String()
	newPod.AddVolumes(config.Spec.Volumes)
	newPod.Status = POD_PENDING_STATUS //此时还未部署,设置状态为Pending
	return newPod
}

func (p Pod) GetPodInfo() ([]byte, error) {
	return ffjson.Marshal(p)
}

func (p *Pod) AddVolumes(volumes []module.Volume) error {
	p.TmpDirMap = make(map[string]string)
	p.HostDirMap = make(map[string]string)
	p.HostFileMap = make(map[string]string)
	for _, value := range volumes {
		if value.Type == emptyDir {
			//临时目录，随机生成
			u := uuid.NewV4()
			path := GetCurrentAbPathByCaller() + "/tmp/" + u.String()
			os.MkdirAll(path, os.ModePerm)
			p.TmpDirMap[value.Name] = path
		} else if value.Type == hostPath {
			//指定了实际目录
			_, err := os.Stat(value.Path)
			if err != nil {
				os.MkdirAll(value.Path, os.ModePerm)
			}
			p.HostDirMap[value.Name] = value.Path
		} else {
			//文件映射
			_, err := os.Stat(value.Path)
			if err != nil {
				return err
			}
			p.HostFileMap[value.Name] = value.Path
		}
	}
	return nil
}

//获取当前文件的路径，
func GetCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
