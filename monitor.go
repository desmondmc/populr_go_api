package main

import "github.com/prometheus/client_golang/prometheus"

var (
	UserCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "populr_user_count",
		Help: "Number of registered users.",
	})
	MessageSentCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "populr_message_sent_count",
		Help: "Number of messages sent.",
	})
	MessageReadCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "populr_message_read_count",
		Help: "Number of messages read.",
	})
)

func initMonitoring() {
	prometheus.MustRegister(UserCount)
	prometheus.MustRegister(MessageSentCount)
	prometheus.MustRegister(MessageReadCount)
}
