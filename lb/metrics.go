package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ps_backend_latency_ms",
		Help:    "Backend latency in ms",
		Buckets: prometheus.LinearBuckets(5, 10, 10),
	}, []string{"backend"})

	metricRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ps_backend_requests_total",
		Help: "Requests proxied per backend",
	}, []string{"backend"})

	metricErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ps_backend_errors_total",
		Help: "Backend errors",
	}, []string{"backend"})
)

func init() {
	prometheus.MustRegister(metricLatency, metricRequests, metricErrors)
}
