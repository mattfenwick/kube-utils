package simulator

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

var eventGauge *prometheus.GaugeVec
var eventCounter *prometheus.CounterVec

func RecordEventValue(name string, tag string, value float64) {
	eventGauge.With(prometheus.Labels{
		"name": name,
		"tag":  tag,
	}).Set(value)
}

func RecordEvent(name string, tag string, err error) {
	eventCounter.With(prometheus.Labels{
		"name":    name,
		"tag":     tag,
		"isError": fmt.Sprintf("%t", err != nil),
	}).Inc()
}

func InitializeMetrics(subsystem string) {
	eventGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kube_utils",
		Subsystem: fmt.Sprintf("simulator_%s", subsystem),
		Name:      "event_gauge",
		Help:      "TODO",
	}, []string{"name", "tag"})
	prometheus.MustRegister(eventGauge)

	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kube_utils",
		Subsystem: fmt.Sprintf("simulator_%s", subsystem),
		Name:      "event_counter",
		Help:      "TODO",
	}, []string{"name", "tag", "isError"})
	prometheus.MustRegister(eventCounter)
}
