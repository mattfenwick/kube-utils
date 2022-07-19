#!/usr/bin/env bash

set -xv
set -euo pipefail

KUBE_VERSION=${KUBE_VERSION:-"1.24.2"}

mkdir -p ./swagger-specs

curl "https://raw.githubusercontent.com/kubernetes/kubernetes/${KUBE_VERSION}/api/openapi-spec/swagger.json" \
  > "./swagger-specs/${KUBE_VERSION}-swagger-spec.json"
