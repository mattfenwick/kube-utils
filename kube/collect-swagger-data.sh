#!/usr/bin/env bash

set -xv
set -euo pipefail

declare -a kube_versions=("v1.16.15" "v1.17.17" "v1.18.19" "v1.19.11" "v1.20.7" "v1.21.2" "v1.22.4" "v1.23.0")

for version in "${kube_versions[@]}"
do
  curl "https://raw.githubusercontent.com/kubernetes/kubernetes/${version}/api/openapi-spec/swagger.json" > "./swagger-specs/${version}-swagger-spec.json"
done
