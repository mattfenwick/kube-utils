#!/usr/bin/env bash

set -xv
set -euo pipefail


## swagger

# compare
go run cmd/schema/main.go compare \
  --version 1.18.19,1.23.0 \
  --type CustomResourceDefinition #> compare-crd.txt

#git diff --no-index old-compare-crd.txt compare-crd.txt

# another compare
go run cmd/schema/main.go compare \
  --version 1.18.0,1.24.2 \
  --type NetworkPolicy,Ingress

# compare-latest
go run cmd/schema/main.go compare-latest

# explain
go run cmd/schema/main.go explain \
  --version 1.18.19 \
  --type CustomResourceDefinition #> explain-crd.txt

#git diff --no-index old-explain-crd.txt explain-crd.txt
