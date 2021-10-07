#!/usr/bin/env bash
set -e

make clean-deployment
make deploy-targets

cd "$(dirname "$0")" # cd to deploy-local.sh is in

function print_msg_and_exit() {
    echo -e "$1"
    exit 1
}

NAMESPACE=wavefront-collector
ROOT_DIR=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)
VERSION=$(cat ./VERSION) #version you want to test
CURRENT_VERSION=${CURRENT_VERSION:-$VERSION}
COLLECTOR_REPO=projects.registry.vmware.com/tanzu_observability/kubernetes-collector
CURRENT_COLLECTOR_REPO=${CURRENT_COLLECTOR_REPO:-$COLLECTOR_REPO}

WF_CLUSTER=nimba
K8S_CLUSTER=$(whoami)-${CURRENT_VERSION}-release-test

echo "Using cluster name '$K8S_CLUSTER' in '$WF_CLUSTER'"

if [[ -z ${WAVEFRONT_TOKEN} ]] ; then
    print_msg_and_exit "wavefront token required"
fi

echo "Temp dir: $TEMP_DIR"

cp "$ROOT_DIR"/deploy/kubernetes/*  "$TEMP_DIR/."
rm "$TEMP_DIR"/kustomization.yaml

cp "$ROOT_DIR/hack/deploy/memcached-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/deploy/mysql-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/deploy/prom-example.yaml" "$TEMP_DIR/."

pushd "$TEMP_DIR"
  sed -i '' "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" "$TEMP_DIR/6-wavefront-proxy.yaml"
  sed -i '' "s/k8s-cluster/${K8S_CLUSTER}/g" "$TEMP_DIR/4-collector-config.yaml"
  sed -i '' "s/wavefront-proxy.default/wavefront-proxy.${NAMESPACE}/g" "$TEMP_DIR/4-collector-config.yaml"

  echo "using version ${CURRENT_VERSION}"
  sed -i '' "s/$VERSION/$CURRENT_VERSION/g" "$TEMP_DIR/5-collector-daemonset.yaml"
  sed -i '' "s%${COLLECTOR_REPO}%${CURRENT_COLLECTOR_REPO}%g" "$TEMP_DIR/5-collector-daemonset.yaml"

  kubectl config set-context --current --namespace="$NAMESPACE"
  kubectl apply -f "$TEMP_DIR/."
  kubectl config set-context --current --namespace=default
popd