package main

import (
	"context"
	"fmt"
	"github.com/mattfenwick/kube-utils/pkg/simulator"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	logrus.Infof("setting up prometheus")
	setupPrometheus(os.Args[1])

	// set up jaeger
	jaegerURL := "otel-collector.monitoring:14268"
	jaegerTracerProvider, err := simulator.NewJaegerTracerProvider(os.Args[1], fmt.Sprintf("http://%s/api/traces", jaegerURL))
	utils.DoOrDie(err)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(jaegerTracerProvider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		utils.DoOrDie(jaegerTracerProvider.Shutdown(ctx))
	}(ctx)

	if os.Args[1] == "server" {
		simulator.RunServer(jaegerTracerProvider)
	} else {
		runClient(jaegerTracerProvider)
	}
}

func runClient(jaegerTracerProvider *tracesdk.TracerProvider) {
	workers := 3
	serverAddress := os.Args[2]
	workers, err := strconv.Atoi(os.Args[3])
	utils.DoOrDie(errors.Wrapf(err, "unable to parse cli arg '%s' to int", os.Args[3]))
	logrus.Infof("server address: %s", serverAddress)
	simulator.RunClient(serverAddress, workers, jaegerTracerProvider)
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
