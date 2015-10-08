package main

import "github.com/prometheus/client_golang/prometheus"

var (
	UserCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "populr_user_count",
		Help: "Number of registered users.",
	})
)

func initMonitoring() {
	prometheus.MustRegister(UserCount)
}
