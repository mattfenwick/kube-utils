#!/usr/bin/env bash

set -xv
set -euo pipefail

docker run -v "$(pwd)/test-values.yaml":/app/val.yaml mfenwick100/py-kube-utils:latest val.yaml
