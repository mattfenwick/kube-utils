#!/usr/bin/env bash

set -xv
set -euo pipefail


## swagger

# compare
go run cmd/api-inspector/main.go swagger compare \
  --version 1.18.19,1.23.0 \
  --type CustomResourceDefinition #> compare-crd.txt

#git diff --no-index old-compare-crd.txt compare-crd.txt

# another compare
go run cmd/api-inspector/main.go swagger compare \
  --version 1.18.0,1.24.2 \
  --type NetworkPolicy,Ingress

# compare-latest
go run cmd/api-inspector/main.go swagger compare-latest

# explain
go run cmd/api-inspector/main.go swagger explain \
  --version 1.18.19 \
  --type CustomResourceDefinition #> explain-crd.txt

#git diff --no-index old-explain-crd.txt explain-crd.txt

# parse
go run cmd/api-inspector/main.go swagger parse --version 1.18.19
