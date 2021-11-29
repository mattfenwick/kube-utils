#!/usr/bin/env bash

set -xv
set -euo pipefail

DOCKERHUB_TAG=mfenwick100/kube-utils/simulator:latest
LOCAL_TAG=localhost:5000/kube-utils/simulator:latest

# build go binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

# build, tag, push docker image
docker build -t $DOCKERHUB_TAG .
docker tag $DOCKERHUB_TAG $LOCAL_TAG
docker push $LOCAL_TAG
