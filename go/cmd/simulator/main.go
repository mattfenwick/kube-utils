package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	logrus.Infof("setting up prometheus")
	setupPrometheus(os.Args[1])

	if os.Args[1] == "server" {
		server()
	} else {
		client()
	}
}

func server() {
	simulator.RunServer()
}

func client() {
	serverAddress := "http://localhost:19999"
	if len(os.Args) >= 3 {
		serverAddress = os.Args[2]
	}
	logrus.Infof("server address: %s", serverAddress)
	simulator.RunClient(serverAddress)
}

func setupPrometheus(subsystemName string) {
	addr := ":9090"

	simulator.InitializeMetrics(subsystemName)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	go func() {
		logrus.Fatal(http.ListenAndServe(addr, nil))
	}()
}
