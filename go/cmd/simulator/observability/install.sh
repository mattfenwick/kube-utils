#!/usr/bin/env bash

set -xv
set -euo pipefail


## Install the Ambassador Edge Stack
kubectl apply -f 1-aes-crds.yaml && \
kubectl wait --for condition=established --timeout=90s crd -lproduct=aes && \
kubectl apply -f 2-aes.yaml && \
kubectl -n ambassador wait --for condition=available --timeout=90s deploy -lproduct=aes


## Install Your Custom Resources
kubectl apply -f 3-user.yaml

## Use and share your Kubernetes playground
# Navigate to https://$AMBASSADOR_SERVICE_IP/ and start exploring your Kubernetes environment!

## Installing Prometheus
kubectl apply -f 4-prometheus-crd.yaml && \
kubectl wait --for condition=established  --timeout=90s crd -lproduct=aes-prometheus && \
kubectl apply -f 5-prometheus.yaml

# The Prometheus web UI will be available at https://$AMBASSADOR_SERVICE_IP/prometheus/

## Installing OpenTelemetry
kubectl apply -f 8-opentelemetry.yaml &&
kubectl -n monitoring wait --for condition=available --timeout=90s deploy -lapp=opentelemetry

#Configure your custom application for B3 Header Propagation and send trace data to one of the OpenTelemetry service endpoint:
#  `otel-collector.monitoring:9411` for Zipkin
#  `otel-collector.monitoring:55680` for OpenTelemetry Protocol (OTLP)
#  `otel-collector.monitoring:14250` for Jaeger-grpc
#  `otel-collector.monitoring:14268` for Jaeger-thrift

## Installing Jaeger
kubectl apply -f 6-jaeger-crds.yaml &&
kubectl apply -f 7-jaeger.yaml &&
kubectl -n monitoring wait --for condition=available --timeout=90s deploy -lapp=jaeger

# The Jaeger UI will be available at https://$AMBASSADOR_SERVICE_IP/jaeger/
