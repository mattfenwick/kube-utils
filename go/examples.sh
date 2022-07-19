#!/usr/bin/env bash

set -xv
set -euo pipefail


# compare
go run cmd/json-finder/main.go compare \
  --version 1.18.19,1.23.0 \
  --type CustomResourceDefinition > compare-crd.txt

git diff --no-index old-compare-crd.txt compare-crd.txt

go run cmd/json-finder/main.go compare \
  --version 1.18.0,1.24.2 \
  --type NetworkPolicy,Ingress


# explain
go run cmd/json-finder/main.go explain \
  --version 1.18.19 \
  --type CustomResourceDefinition > explain-crd.txt

git diff --no-index old-explain-crd.txt explain-crd.txt


# parse
go run cmd/json-finder/main.go parse --version 1.18.19
