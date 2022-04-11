package dockerClient

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"minik8s/cmd/kubelet/app/message"
	"minik8s/cmd/kubelet/app/module"
	"sync"
	"unsafe"
)

type dockerClient struct {
	cli *client.Client
}

var instance *dockerClient
var lock sync.Mutex

func getInstance() *dockerClient {
	//lock.Lock()
	//defer lock.Unlock()
	if instance == nil {
		instance = new(dockerClient)
		cli, err := client.NewClientWithOpts()
		instance.cli = cli
		if err != nil {
			//may quit here
			panic(err)
			return nil
		}
		return instance
	} else {
		return instance
	}
}

//获取所有容器,docker ps -a
func getAllContainers() ([]types.Container, error) {
	cli := getInstance().cli
	return cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
}
func getRunningContainers() ([]types.Container, error) {
	cli := getInstance().cli
	return cli.ContainerList(context.Background(), types.ContainerListOptions{})
}
func startContainer(containerId string) error {
	cli := getInstance().cli
	err := cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{})
	return err
}
func stopContainer(containerId string) error {
	cli := getInstance().cli
	err := cli.ContainerStop(context.Background(), containerId, nil)
	return err
}
func getPodNetworkSettings(containerId string) (*types.NetworkSettings, error) {
	cli := getInstance().cli
	res, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return nil, err
	}
	return res.NetworkSettings, nil
}
func dockerClientPullImages(images []string) error {
	for _, value := range images {
		err := dockerClientPullSingleImage(value)
		if err != nil {
			return err
		}
	}
	return nil
}

//注意， 调用ImagePull 函数， 拉取进程在后台运行，因此要保证前台挂起足够时间保证拉取成功
func dockerClientPullSingleImage(image string) error {
	cli := getInstance().cli
	out, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}
func createContainersOfPod(containers []module.Container) ([]string, error) {
	cli := getInstance().cli
	flag := true
	var firstContainerId string
	var result []string
	for _, value := range containers {
		if flag {
			flag = false
			//生成开放端口
			var exports nat.PortSet
			if value.Ports != nil {
				exports := make(nat.PortSet, len(value.Ports))
				for _, port := range value.Ports {
					p, err := nat.NewPort("tcp", port.ContainerPort)
					if err != nil {
						return nil, err
					}
					exports[p] = struct{}{}
				}
			}
			err := dockerClientPullSingleImage(value.Image)
			if err != nil {
				return nil, err
			}
			var mounts []mount.Mount
			if value.VolumeMounts != nil {
				for _, it := range value.VolumeMounts {
					mounts = append(mounts, mount.Mount{
						Type:   mount.TypeBind,
						Source: it.Name,
						Target: it.MountPath,
					})
				}
			}

			resp, err := cli.ContainerCreate(context.Background(), &container.Config{
				Image:        value.Image,
				ExposedPorts: exports,
				Cmd:          value.Command,
			}, &container.HostConfig{
				Mounts: mounts,
			}, nil, nil, value.Name)
			if err != nil {
				return nil, err
			}
			firstContainerId = resp.ID
			result = append(result, firstContainerId)
		} else {
			//只有第一个container可以生成开放端口
			//先拉取镜像
			err := dockerClientPullSingleImage(value.Image)
			if err != nil {
				return nil, err
			}
			var mounts []mount.Mount
			if value.VolumeMounts != nil {
				for _, it := range value.VolumeMounts {
					mounts = append(mounts, mount.Mount{
						Type:   mount.TypeBind,
						Source: it.Name,
						Target: it.MountPath,
					})
				}
			}
			resp, err := cli.ContainerCreate(context.Background(), &container.Config{
				Image: value.Image,
				Cmd:   value.Command,
			}, &container.HostConfig{
				NetworkMode: container.NetworkMode("container:" + firstContainerId),
				Mounts:      mounts,
			}, nil, nil, value.Name)
			if err != nil {
				return nil, err
			}
			result = append(result, resp.ID)
		}
	}
	return result, nil
}

//涉及大量指针操作，要确保在caller和callee在同一个地址空间中
func HandleCommand(command *message.Command) *message.Response {

	switch command.CommandType {
	case message.COMMAND_GET_ALL_CONTAINER:
		containers, err := getAllContainers()
		var result message.ResponseWithContainInfo
		result.Response.CommandType = message.COMMAND_GET_ALL_CONTAINER
		result.Response.Err = err
		result.Containers = containers
		return &(result.Response)
	case message.COMMAND_GET_RUNNING_CONTAINER:
		containers, err := getRunningContainers()
		var result message.ResponseWithContainInfo
		result.Response.CommandType = message.COMMAND_GET_RUNNING_CONTAINER
		result.Response.Err = err
		result.Containers = containers
		return &(result.Response)
	case message.COMMAND_RUN_CONTAINER:
		p := (*message.CommandWithId)(unsafe.Pointer(command))
		err := startContainer(p.ContainerId)
		var result message.Response
		result.CommandType = message.COMMAND_RUN_CONTAINER
		result.Err = err
		return &result
	case message.COMMAND_STOP_CONTAINER:
		p := (*message.CommandWithId)(unsafe.Pointer(command))
		err := stopContainer(p.ContainerId)
		var result message.Response
		result.CommandType = message.COMMAND_STOP_CONTAINER
		result.Err = err
		return &result
	case message.COMMAND_BUILD_CONTAINERS_OF_POD:
		p := (*message.CommandWithConfig)(unsafe.Pointer(command))
		res, err := createContainersOfPod(p.Group)
		var result message.ResponseWithContainIds
		result.Err = err
		result.CommandType = message.COMMAND_BUILD_CONTAINERS_OF_POD
		result.ContainersIds = res
		return &(result.Response)
	case message.COMMAND_PULL_IMAGES:
		p := (*message.CommandWithImages)(unsafe.Pointer(command))
		err := dockerClientPullImages(p.Images)
		var result message.Response
		result.CommandType = message.COMMAND_PULL_IMAGES
		result.Err = err
		return &result
	}
	return nil
}
