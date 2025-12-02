package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metricsHandler returns an http.Handler that serves Prometheus metrics.
func NewMetricsHandler() http.Handler {
	return promhttp.Handler()
}

var HttpRequestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	},
	[]string{"path", "method"}, // Labels for path and method
)

func init() {
	// Register the counter with the default Prometheus registry
	prometheus.MustRegister(HttpRequestCounter)
}
