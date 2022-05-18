package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	podMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_monitor",
	}, []string{"resource", "pod", "uid", "selector"})
)

func MakeMetricRecord(podName string, podUID string, selectorPair *string, memPercent float64, cpuPercent float64) {
	podTag := podName
	if selectorPair == nil {
		podMetric.WithLabelValues("memory", podTag, podUID, "").Set(memPercent)
		podMetric.WithLabelValues("cpu", podTag, podUID, "").Set(cpuPercent)
	} else {
		podMetric.WithLabelValues("memory", podTag, podUID, *selectorPair).Set(memPercent)
		podMetric.WithLabelValues("cpu", podTag, podUID, *selectorPair).Set(cpuPercent)
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
