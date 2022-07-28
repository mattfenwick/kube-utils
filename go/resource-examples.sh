#!/usr/bin/env bash

set -xv
set -euo pipefail


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
