package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
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
	workers := 3
	var err error
	if len(os.Args) >= 4 {
		serverAddress = os.Args[2]
		workers, err = strconv.Atoi(os.Args[3])
		utils.DoOrDie(errors.Wrapf(err, "unable to parse cli arg '%s' to int", os.Args[3]))
	}
	logrus.Infof("server address: %s", serverAddress)
	simulator.RunClient(serverAddress, workers)
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
