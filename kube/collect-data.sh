#!/usr/bin/env bash

set -xv
set -euo pipefail

declare -a kind_images=("v1.18.19" "v1.19.11" "v1.20.7" "v1.21.2" "v1.22.4" "v1.23.0")

for version in "${kind_images[@]}"
do
  name="api-resources-$version"
  kind create cluster --image="kindest/node:$version" --name="$name"
  kubectl api-resources -o wide > data/"$version"-api-resources.txt
  kubectl api-versions > data/"$version"-api-versions.txt
  kind delete cluster --name="$name"
done
