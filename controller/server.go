package controller

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startServer starts a server for metrics and health checking.
func (c *controller) startServer() {
	log.Println("Starting server on:", c.listenAddress)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(c.listenAddress, nil))
}
