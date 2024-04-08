package metrics

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	ServiceColdstartCount *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := metrics{
		ServiceColdstartCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_start_count",
			Help: "Service Cold/Warm start Count",
		}, []string{"service_name", "service_namespace", "cold_hot"}),
	}
	reg.MustRegister(m.ServiceColdstartCount)
	return &m
}
