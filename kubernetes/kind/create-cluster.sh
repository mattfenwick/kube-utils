#!/bin/bash

set -xv
set -euo pipefail

KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-"kube-utils"}
KIND_NODE_IMAGE=${KIND_NODE_IMAGE:-"kindest/node:v1.20.7"}
REGISTRY_IMAGE=${REGISTRY_IMAGE:-"docker.io/library/registry:2"}
REGISTRY_NAME=${REGISTRY_NAME:-'kind-registry'}
REGISTRY_PORT=${REGISTRY_PORT:-'5000'}
METRICS_NS=${METRICS_NS:-"prometheus"}

running="$(docker inspect -f '{{.State.Running}}' "${REGISTRY_NAME}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${REGISTRY_PORT}:5000" --name "${REGISTRY_NAME}" \
    "$REGISTRY_IMAGE"
fi

kind create cluster \
  --name "$KIND_CLUSTER_NAME" \
  --config config.yaml \
  --image "$KIND_NODE_IMAGE"

docker network connect "kind" "${REGISTRY_NAME}" || true

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<-EOF  | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: local-registry-hosting
      namespace: kube-public
    data:
      localRegistryHosting.v1: |
        host: "localhost:${REGISTRY_PORT}"
        help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

kubectl get nodes
kubectl wait --for=condition="Ready" nodes --all --timeout="15m"


# set up ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl get pods -A
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=10m


# set up metrics and prometheus
kubectl apply -f ./metrics-server.yaml

if ! helm repo list | grep -i prometheus-community; then
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
fi

kubectl create ns "$METRICS_NS"
helm upgrade --install my-prom prometheus-community/prometheus \
  --debug \
  --version 14.0.0 \
  --namespace "$METRICS_NS"
