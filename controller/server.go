package controller

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startServer starts a server for metrics and health checking.
func (c *controller) startServer() {
	log.Println("starting server on:", c.config.ListenAddress)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(c.config.ListenAddress, nil))
}
