#!/usr/bin/env bash

set -xv
set -euo pipefail


# parse
go run cmd/api-inspector/main.go swagger-debug parse --version 1.18.19

# go run cmd/api-inspector/main.go swagger-debug analyze-schema

# go run cmd/api-inspector/main.go swagger-debug test-schema-parser
