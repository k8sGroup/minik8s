package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	podMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_monitor",
	}, []string{"resource", "node", "pod", "selector"})
)

func MakeMetricRecord(nodeName string, podName string, selectorPair *string, memPercent float64, cpuPercent float64) {
	nodeTag := nodeName
	podTag := podName
	if selectorPair == nil {
		podMetric.WithLabelValues("memory", nodeTag, podTag, "").Set(memPercent)
		podMetric.WithLabelValues("cpu", nodeTag, podTag, "").Set(cpuPercent)
	} else {
		podMetric.WithLabelValues("memory", nodeTag, podTag, *selectorPair).Set(memPercent)
		podMetric.WithLabelValues("cpu", nodeTag, podTag, *selectorPair).Set(cpuPercent)
	}
}

func selectorAppTag(label map[string]string) *string {
	for key, value := range label {
		if key == "app" {
			return &value
		}
	}
	return nil
}
