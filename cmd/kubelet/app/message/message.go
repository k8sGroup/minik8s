package message

import (
	"github.com/docker/docker/api/types"
	"minik8s/cmd/kubelet/app/module"
)

//---------------------------Container Part---------------------------------------//
const COMMAND_GET_ALL_CONTAINER = 0
const COMMAND_GET_RUNNING_CONTAINER = 1
const COMMAND_RUN_CONTAINER = 2
const COMMAND_STOP_CONTAINER = 3
const COMMAND_BUILD_CONTAINERS_OF_POD = 4
const COMMAND_PULL_IMAGES = 5

//------------------------------------------------------------------------------//

//-------------------------- Pod Part ------------------------------------------//
const ADD_POD = 0

const DELETE_POD = 1

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
	Group []module.Container
}

type CommandWithImages struct {
	Command
	Images []string
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
	ContainersIds []string
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
