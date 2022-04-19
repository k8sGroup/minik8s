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
	"unsafe"
)

func getNewClient() (*client.Client, error) {
	return client.NewClientWithOpts()
}

//获取所有容器,docker ps -a
func getAllContainers() ([]types.Container, error) {
	cli, err := getNewClient()
	if err != nil {
		return nil, err
	}
	return cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
}
func getRunningContainers() ([]types.Container, error) {
	cli, err := getNewClient()
	if err != nil {
		return nil, err
	}
	return cli.ContainerList(context.Background(), types.ContainerListOptions{})
}
func startContainer(containerId string) error {
	cli, err := getNewClient()
	if err != nil {
		return err
	}
	err = cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{})
	return err
}
func stopContainer(containerId string) error {
	cli, err := getNewClient()
	if err != nil {
		return err
	}
	err = cli.ContainerStop(context.Background(), containerId, nil)
	return err
}
func getPodNetworkSettings(containerId string) (*types.NetworkSettings, error) {
	cli, err := getNewClient()
	if err != nil {
		return nil, err
	}
	res, err2 := cli.ContainerInspect(context.Background(), containerId)
	if err2 != nil {
		return nil, err
	}
	return res.NetworkSettings, nil
}
func isImageExist(a string, b string) bool {
	if a == b {
		return true
	}
	tmp := a + ":latest"
	if tmp == b {
		return true
	}
	return false
}
func dockerClientPullImages(images []string) error {
	//先统一拉取镜像，确认是否已经存在于本地
	cli, err2 := getNewClient()
	if err2 != nil {
		return err2
	}
	resp, err := cli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return err
	}
	var filter []string
	for _, value := range images {
		flag := false
		for _, it := range resp {
			if isImageExist(value, it.RepoTags[0]) {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		filter = append(filter, value)
	}
	if filter != nil {
		for _, value := range filter {
			err := dockerClientPullSingleImage(value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//注意， 调用ImagePull 函数， 拉取进程在后台运行，因此要保证前台挂起足够时间保证拉取成功
func dockerClientPullSingleImage(image string) error {
	cli, err2 := getNewClient()
	if err2 != nil {
		return err2
	}
	out, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}
func runContainers(containerIds []module.ContainerMeta) error {
	cli, err2 := getNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range containerIds {
		err := cli.ContainerStart(context.Background(), value.ContainerId, types.ContainerStartOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
func getContainersInfo(containerIds []string) ([]types.ContainerJSON, error) {
	cli, err2 := getNewClient()
	if err2 != nil {
		return nil, err2
	}
	var result []types.ContainerJSON
	for _, value := range containerIds {
		single, err := cli.ContainerInspect(context.Background(), value)
		if err != nil {
			return nil, err
		}
		result = append(result, single)
	}
	return result, nil
}

//创建pause容器
func createPause(ports []module.Port, name string) (container.ContainerCreateCreatedBody, error) {
	cli, err2 := getNewClient()
	if err2 != nil {
		return container.ContainerCreateCreatedBody{}, err2
	}
	var exports nat.PortSet
	exports = make(nat.PortSet, len(ports))
	for _, port := range ports {
		p, err := nat.NewPort("tcp", port.ContainerPort)
		if err != nil {
			return container.ContainerCreateCreatedBody{}, err
		}
		exports[p] = struct{}{}
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image:        "gcr.io/google_containers/pause-amd64:3.0",
		ExposedPorts: exports,
	}, &container.HostConfig{
		IpcMode: container.IpcMode("shareable"),
		PortBindings: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8080",
				},
			},
		},
	}, nil, nil, name)
	return resp, err
}

//检查容器状态
func probeContainers(containerIds []string) ([]string, error) {
	cli, err2 := getNewClient()
	if err2 != nil {
		return nil, err2
	}
	var res []string
	for _, value := range containerIds {
		resp, err := cli.ContainerInspect(context.Background(), value)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.State.Status)
	}
	return res, nil
}

//删除containers
func deleteContainers(containerIds []string) error {
	cli, err2 := getNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range containerIds {
		err := cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

//查找是否存在，存在就删除
func deleteExitedContainers(names []string) error {
	cli, err2 := getNewClient()
	if err2 != nil {
		return err2
	}
	for _, value := range names {
		_, err := cli.ContainerInspect(context.Background(), value)
		if err == nil {
			err = cli.ContainerRemove(context.Background(), value, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func createContainersOfPod(containers []module.Container) ([]module.ContainerMeta, error) {
	cli, err2 := getNewClient()
	if err2 != nil {
		return nil, err2
	}
	var firstContainerId string
	var result []module.ContainerMeta
	//先生成所有要暴露的port集合
	var totlePort []module.Port
	images := []string{"gcr.io/google_containers/pause-amd64:3.0"}
	//防止重名，先检查是否重名，有的话删除
	var names []string
	pauseName := "pause"
	for _, value := range containers {
		pauseName += "_" + value.Name
		names = append(names, value.Name)
		images = append(images, value.Image)
		for _, port := range value.Ports {
			totlePort = append(totlePort, port)
		}
	}
	names = append(names, pauseName)
	err3 := deleteExitedContainers(names)
	if err3 != nil {
		return nil, err3
	}
	//先统一拉取镜像
	err := dockerClientPullImages(images)
	if err != nil {
		return nil, err
	}
	//创建pause容器
	pause, err := createPause(totlePort, pauseName)
	if err != nil {
		return nil, err
	}
	firstContainerId = pause.ID
	result = append(result, module.ContainerMeta{
		RealName:    pauseName,
		ContainerId: firstContainerId,
	})
	for _, value := range containers {
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
		//生成env
		var env []string
		if value.Env != nil {
			for _, it := range value.Env {
				singleEnv := it.Name + "=" + it.Value
				env = append(env, singleEnv)
			}
		}
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image:      value.Image,
			Entrypoint: value.Command,
			Cmd:        value.Args,
			Env:        env,
		}, &container.HostConfig{
			NetworkMode: container.NetworkMode("container:" + firstContainerId),
			Mounts:      mounts,
			IpcMode:     container.IpcMode("container:" + firstContainerId),
			PidMode:     container.PidMode("container" + firstContainerId),
		}, nil, nil, value.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, module.ContainerMeta{
			RealName:    value.Name,
			ContainerId: resp.ID,
		})
	}
	//启动容器
	err = runContainers(result)
	if err != nil {
		return nil, err
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
		result.Containers = res
		return &(result.Response)
	case message.COMMAND_PULL_IMAGES:
		p := (*message.CommandWithImages)(unsafe.Pointer(command))
		err := dockerClientPullImages(p.Images)
		var result message.Response
		result.CommandType = message.COMMAND_PULL_IMAGES
		result.Err = err
		return &result
	case message.COMMAND_PROBE_CONTAINER:
		p := (*message.CommandWithContainerIds)(unsafe.Pointer(command))
		res, err := probeContainers(p.ContainerIds)
		var result message.ResponseWithProbeInfos
		result.Err = err
		result.CommandType = message.COMMAND_PROBE_CONTAINER
		result.ProbeInfos = res
		return &(result.Response)
	case message.COMMAND_DELETE_CONTAINER:
		//删除containers的操作
		p := (*message.CommandWithContainerIds)(unsafe.Pointer(command))
		err := deleteContainers(p.ContainerIds)
		var result message.Response
		result.CommandType = message.COMMAND_DELETE_CONTAINER
		result.Err = err
		return &result
	}
	return nil
}
