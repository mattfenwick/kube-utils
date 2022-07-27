#!/usr/bin/env bash

set -xv
set -euo pipefail


## swagger

# resource: compare
go run cmd/schema/main.go resource compare \
  --version 1.18.19,1.24.3 \
  --type CustomResourceDefinition,CronJob #> compare-crd.txt

go run cmd/schema/main.go resource compare \
  --version 1.18.0,1.24.2 \
  --type NetworkPolicy,Ingress

# resource: explain
go run cmd/schema/main.go resource explain \
  --version 1.18.19 \
  --type CustomResourceDefinition #> explain-crd.txt

go run cmd/schema/main.go resource explain \
  --version 1.18.19 \
  --type CronJob > cronjob-1-18.txt

go run cmd/schema/main.go resource explain \
  --version 1.24.3 \
  --type CronJob > cronjob-1-24.txt

# gvk: explain
KUBE_VERSIONS=1.18.20,1.20.15,1.22.12,1.24.0,1.25.0-alpha.3
KUBE_RESOURCES=Ingress,CronJob,CustomResourceDefinition
go run cmd/schema/main.go gvk explain \
  --resource="$KUBE_RESOURCES" \
  --kube-version="$KUBE_VERSIONS"

go run cmd/schema/main.go gvk explain \
  --resource="$KUBE_RESOURCES" \
  --kube-version="$KUBE_VERSIONS" \
  --group-by=api-version

go run cmd/schema/main.go gvk explain \
  --resource="$KUBE_RESOURCES" \
  --kube-version="$KUBE_VERSIONS" \
  --diff

go run cmd/schema/main.go gvk explain \
  --resource="$KUBE_RESOURCES" \
  --kube-version="$KUBE_VERSIONS" \
  --group-by=api-version \
  --diff