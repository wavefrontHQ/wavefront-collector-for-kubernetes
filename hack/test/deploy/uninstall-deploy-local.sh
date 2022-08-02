#!/usr/bin/env bash
set -e

cd "$(dirname "$0")" # cd to deploy-local.sh is in

source "./k8s-utils.sh"

ROOT_DIR=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)
NS=wavefront-collector

cp "$ROOT_DIR"/deploy/kubernetes/*.yaml  "$TEMP_DIR/."
rm "$TEMP_DIR"/kustomization.yaml || true
cp "$ROOT_DIR/hack/test/deploy/memcached-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/mysql-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/prom-example.yaml" "$TEMP_DIR/."

pushd "$TEMP_DIR"
  kubectl config set-context --current --namespace="$NS"
  kubectl delete -f "$TEMP_DIR/."
  kubectl config set-context --current --namespace=default
popd
