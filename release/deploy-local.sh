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
CURRENT_VERSION= #set if you want to test something other than version in the 5-collector-daemonset.yaml
VERSION=$(cat ./VERSION) #version you want to test

WF_CLUSTER=nimba
K8S_CLUSTER=$VERSION-release-test

if [[ -z ${WAVEFRONT_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
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

if [[ -n ${CURRENT_VERSION} ]] ; then
    echo "has current version"
    sed -i '' "s/$CURRENT_VERSION/$VERSION/g" "$TEMP_DIR/5-collector-daemonset.yaml"
fi

  kubectl config set-context --current --namespace="$NAMESPACE"
  kubectl apply -f "$TEMP_DIR/."
  kubectl config set-context --current --namespace=default
popd