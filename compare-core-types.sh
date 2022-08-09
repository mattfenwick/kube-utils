#!/usr/bin/env bash

set -xv
set -euo pipefail

go run cmd/json-finder/main.go compare \
  --version=1.18.19,1.23.0 \
  --type="Service,ClusterRole,ClusterRoleBinding,ConfigMap,CronJob,CustomResourceDefinition,Deployment,Ingress,Job,Role,RoleBinding,Secret,ServiceAccount,StatefulSet"