#!/usr/bin/env bash

set -xv
set -euo pipefail

kubectl create ns simulator || true

kubectl apply -n simulator \
  -f kube-resources.yaml
