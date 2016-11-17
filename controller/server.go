package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	healthCheckRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_check_request_total",
			Help: "Number of health check requests",
		},
		[]string{"success"},
	)
	healthCheckRequestTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "health_check_request_milliseconds",
		Help: "Time taken to respond to health check, in milliseconds",
	})
)

func init() {
	prometheus.MustRegister(healthCheckRequests)
	prometheus.MustRegister(healthCheckRequestTime)
}

// startServer starts a server for metrics and health checking.
func (c *controller) startServer() {
	log.Println("starting server on:", c.config.ListenAddress)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if _, err := c.clientset.Extensions().ThirdPartyResources().Get(ClusterThirdPartyResourceName); err != nil {
			healthCheckRequests.WithLabelValues("failed").Inc()

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			healthCheckRequests.WithLabelValues("successful").Inc()
		}

		healthCheckRequestTime.Set(float64(time.Since(start) / time.Millisecond))
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(c.config.ListenAddress, nil))
}
