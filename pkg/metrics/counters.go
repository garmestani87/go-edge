package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpCall = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "db_calls_total",
		Help: "Number of Service calls",
	}, []string{"path", "method", "status_code"},
)
