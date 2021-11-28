#!/usr/bin/env bash

set -xv
set -euo pipefail

go fmt ./...
go vet ./...
