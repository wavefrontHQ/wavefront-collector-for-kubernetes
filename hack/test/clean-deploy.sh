#!/usr/bin/env bash
set -e

NS=$(kubectl get namespaces | awk '/wavefront-collector/ {print $1}')

if [ -z ${NS} ]; then exit 0; fi

./generate.sh -c "fake" -t "fake" -v "fake" -d "fake"

echo "deleting wavefront collector deployment"
kustomize build base | kubectl delete -f - || true