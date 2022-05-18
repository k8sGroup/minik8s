package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"minik8s/pkg/kubelet/pod"
	"net/http"
)

type IMonitor interface {
}

type DockerMonitor struct {
	dockerClient *client.Client
}

func NewDockerMonitor() *DockerMonitor {
	fmt.Printf("[NewDockerMonitor] Init enter\n")
	c, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Printf("[NewDockerMonitor] Init client fail\n")
		return nil
	}

	fmt.Printf("[NewDockerMonitor] Init client\n")
	return &DockerMonitor{
		dockerClient: c,
	}

}

func (m *DockerMonitor) Listener() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9070", nil))
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
			continue
		}
		raw, _ := ioutil.ReadAll(stats.Body)
		//fmt.Printf("[MetricDockerStat] Stat info:%v\n", string(raw))

		// unmarshal json
		statsJson := &types.StatsJSON{}
		err = json.Unmarshal(raw, statsJson)
		if err != nil {
			fmt.Printf("[MetricDockerStat] Unmarshal fail,err:%v", err)
			continue
		}

		// set metrics
		memPercent := getMemPercent(statsJson)
		cpuPercent := getCPUPercent(statsJson)

		//// set metrics, test function
		//memPercent := m.r.Float64()
		//cpuPercent := m.r.Float64()

		fmt.Printf("[MetricDockerStat] cpu:%f%%  mem:%f%%\n", cpuPercent, memPercent)

		serviceTag := selectorAppTag(pod.Label)
		MakeMetricRecord(pod.GetName(), pod.GetUID(), serviceTag, memPercent, cpuPercent)
	}
}

func getCPUPercent(statsJson *types.StatsJSON) float64 {
	// cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	var preCPUUsage uint64
	var CPUUsage uint64
	preCPUUsage = 0
	CPUUsage = 0
	for _, core := range statsJson.CPUStats.CPUUsage.PercpuUsage {
		CPUUsage += core
	}
	for _, core := range statsJson.PreCPUStats.CPUUsage.PercpuUsage {
		preCPUUsage += core
	}
	systemUsage := statsJson.CPUStats.SystemUsage
	preSystemUsage := statsJson.PreCPUStats.SystemUsage

	deltaCPU := CPUUsage - preCPUUsage
	deltaSystem := systemUsage - preSystemUsage

	onlineCPU := statsJson.CPUStats.OnlineCPUs

	fmt.Printf("deltaCPU:%v deltaSystem:%v onlineCPU:%v\n", deltaCPU, deltaSystem, onlineCPU)

	cpuPercent := (float64(deltaCPU) / float64(deltaSystem)) * float64(onlineCPU) * 100.0
	return cpuPercent
}

func getMemPercent(statsJson *types.StatsJSON) float64 {
	// MEM USAGE / LIMIT
	usage := statsJson.MemoryStats.Usage
	maxUsage := statsJson.MemoryStats.Limit
	percentage := float64(usage) / float64(maxUsage) * 100.0
	return percentage
}
