package message

import (
	"github.com/docker/docker/api/types"
	"minik8s/object"
)

//---------------------------Container Part---------------------------------------//
const COMMAND_GET_ALL_CONTAINER = 0
const COMMAND_GET_RUNNING_CONTAINER = 1
const COMMAND_RUN_CONTAINER = 2
const COMMAND_STOP_CONTAINER = 3
const COMMAND_BUILD_CONTAINERS_OF_POD = 4
const COMMAND_PULL_IMAGES = 5
const COMMAND_PROBE_CONTAINER = 6
const COMMAND_DELETE_CONTAINER = 7

//------------------------------------------------------------------------------//

//-------------------------- Pod Part ------------------------------------------//
const ADD_POD = 0
const DELETE_POD = 1
const PROBE_POD = 2

//-------------------------------------------------------------------------------//
type Command struct {
	CommandType int
}

type CommandWithId struct {
	Command
	ContainerId string
}
type CommandWithConfig struct {
	Command
	Group []object.Container
}

type CommandWithImages struct {
	Command
	Images []string
}

//commandType are COMMAND_PROBE_CONTAINER|COMMAND_DELETE_CONTAINER
type CommandWithContainerIds struct {
	Command
	ContainerIds []string
}

type Response struct {
	CommandType int
	Err         error
}
type ResponseWithContainInfo struct {
	Response
	Containers []types.Container
}

type ResponseWithContainIds struct {
	Response
	Containers   []object.ContainerMeta
	NetWorkInfos *types.NetworkSettings
}

//返回的切片中元素顺序与commnad中容器id顺序一一对应
type ResponseWithProbeInfos struct {
	Response
	ProbeInfos []string
}

type PodCommand struct {
	ContainerCommand *Command
	PodUid           string
	PodCommandType   int
}
type PodResponse struct {
	ContainerResponse *Response
	PodUid            string
	PodResponseType   int
}
