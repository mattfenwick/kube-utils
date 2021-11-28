#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE_TAG=${IMAGE_TAG:-latest}
IMAGE=mfenwick100/py-kube-utils:"$IMAGE_TAG"

docker build . -t "$IMAGE"

#docker push "$IMAGE"
