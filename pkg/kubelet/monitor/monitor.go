package monitor

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io/ioutil"
	"minik8s/pkg/kubelet/pod"
)

type IMonitor interface {
}

type DockerMonitor struct {
	// TODO: store the data that can identify a node
	nodeName     string
	dockerClient *client.Client
}

func NewDockerMonitor() *DockerMonitor {
	c, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Printf("[NewDockerMonitor] Init client fail\n")
		return nil
	}
	InitPrometheus()
	return &DockerMonitor{
		nodeName:     "test",
		dockerClient: c,
	}
}

func (m *DockerMonitor) MetricDockerStat(ctx context.Context, pod *pod.Pod) {
	// due to make deep copy of pod map, it may be released
	if pod == nil {
		return
	}
	containers := pod.GetContainers()
	for _, container := range containers {
		containerID := container.ContainerId
		stats, err := m.dockerClient.ContainerStats(ctx, containerID, false)
		if err != nil {
			fmt.Printf("[MetricDockerStat] Get stats error:%v\n", err)
		}
		raw, _ := ioutil.ReadAll(stats.Body)
		fmt.Printf("[MetricDockerStat] Stat info:%v\n", string(raw))

		// unmarshal json

		// set metrics
		var memPercent, cpuPercent float64
		serviceTag := selectorAppTag(pod.Label)
		MakeMetricRecord(m.nodeName, pod.GetName(), serviceTag, memPercent, cpuPercent)
	}

}
