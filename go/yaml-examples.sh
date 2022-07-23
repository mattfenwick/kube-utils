#!/usr/bin/env bash

set -xv
set -euo pipefail

## yaml
go run cmd/api-inspector/main.go analyze-yaml --path ./example.yaml
