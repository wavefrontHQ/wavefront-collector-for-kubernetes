#!/usr/bin/env bash
set -e

cd "$(dirname "$0")" # cd to deploy-local.sh is in

source "../hack/test/deploy/k8s-utils.sh"

if [[ -z ${WAVEFRONT_TOKEN} ]] ; then
    print_msg_and_exit "WAVEFRONT_TOKEN required"
fi

NS=wavefront-collector
ROOT_DIR=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)
VERSION=$(cat ./VERSION) #version you want to test
CURRENT_VERSION=${CURRENT_VERSION:-$VERSION}
COLLECTOR_REPO=projects.registry.vmware.com/tanzu_observability/kubernetes-collector
CURRENT_COLLECTOR_REPO=${CURRENT_COLLECTOR_REPO:-$COLLECTOR_REPO}

pushd ../
  make clean-deployment
  make deploy-targets
popd

if [[ -z ${WF_CLUSTER} ]] ; then
    WF_CLUSTER=nimba
fi

if [[ -z ${CONFIG_CLUSTER_NAME} ]] ; then
    CONFIG_CLUSTER_NAME=$(whoami)-${CURRENT_VERSION}-release-test
fi

echo "Using cluster name '$CONFIG_CLUSTER_NAME' in '$WF_CLUSTER'"
echo "Temp dir: $TEMP_DIR"

cp "$ROOT_DIR"/deploy/kubernetes/*  "$TEMP_DIR/."
rm "$TEMP_DIR"/kustomization.yaml

cp "$ROOT_DIR/hack/test/deploy/memcached-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/mysql-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/prom-example.yaml" "$TEMP_DIR/."

pushd "$TEMP_DIR"
  sed -i "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" "$TEMP_DIR/6-wavefront-proxy.yaml"
  sed -i "s/k8s-cluster/${CONFIG_CLUSTER_NAME}/g" "$TEMP_DIR/4-collector-config.yaml"

  echo "using version ${CURRENT_VERSION}"
  sed -i "s/$VERSION/$CURRENT_VERSION/g" "$TEMP_DIR/5-collector-daemonset.yaml"
  sed -i "s%${COLLECTOR_REPO}%${CURRENT_COLLECTOR_REPO}%g" "$TEMP_DIR/5-collector-daemonset.yaml"

  kubectl config set-context --current --namespace="$NS"
  kubectl apply -f "$TEMP_DIR/."
  kubectl config set-context --current --namespace=default
popd
