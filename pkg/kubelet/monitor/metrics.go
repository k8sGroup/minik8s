package monitor

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func InitPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(podMetric)
}

var (
	podMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_monitor",
	}, []string{})
)

func MakeMetricRecord(nodeName string, podName string, selectorPair *string, memPercent float64, cpuPercent float64) {

	nodeTag := fmt.Sprintf("node=%s", nodeName)
	podTag := fmt.Sprintf("pod=%s", podName)
	if selectorPair == nil {
		podMetric.WithLabelValues("memory", nodeTag, podTag).Set(memPercent)
		podMetric.WithLabelValues("cpu", nodeTag, podTag).Set(cpuPercent)
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
