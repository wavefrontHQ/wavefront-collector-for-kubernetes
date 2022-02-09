#!/usr/bin/env bash
set -e

cd "$(dirname "$0")" # cd to deploy-local.sh is in

source "./k8s-utils.sh"

if [[ -z ${WAVEFRONT_TOKEN} ]] ; then
    print_msg_and_exit "WAVEFRONT_TOKEN required"
fi

NS=wavefront-collector
ROOT_DIR=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)

# TODO: this rename breaks Jenkins so if we go with it we have updating to do
DEFAULT_COLLECTOR_VERSION=$(cat ${ROOT_DIR}/release/VERSION) #version you want to test
CURRENT_COLLECTOR_VERSION=${CURRENT_COLLECTOR_VERSION:-$DEFAULT_COLLECTOR_VERSION}
DEFAULT_COLLECTOR_REPO=projects.registry.vmware.com/tanzu_observability/kubernetes-collector
CURRENT_COLLECTOR_REPO=${CURRENT_COLLECTOR_REPO:-$DEFAULT_COLLECTOR_REPO}

DEFAULT_PROXY_VERSION=10.11
CURRENT_PROXY_VERSION=${CURRENT_PROXY_VERSION:-$DEFAULT_PROXY_VERSION}
DEFAULT_PROXY_REPO=projects.registry.vmware.com/tanzu_observability/proxy
CURRENT_PROXY_REPO=${CURRENT_PROXY_REPO:-$DEFAULT_PROXY_REPO}

pushd "$ROOT_DIR"
  make clean-deployment
  make deploy-targets
popd

if [[ -z ${WF_CLUSTER} ]] ; then
    WF_CLUSTER=nimba
fi

if [[ -z ${CONFIG_CLUSTER_NAME} ]] ; then
    CONFIG_CLUSTER_NAME=$(whoami)-${CURRENT_COLLECTOR_VERSION}-release-test
fi

echo "Using cluster name '$CONFIG_CLUSTER_NAME' in '$WF_CLUSTER'"
echo "Temp dir: $TEMP_DIR"

cp "$ROOT_DIR"/deploy/kubernetes/*  "$TEMP_DIR/."
rm "$TEMP_DIR"/kustomization.yaml

cp "$ROOT_DIR/hack/test/deploy/memcached-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/mysql-config.yaml" "$TEMP_DIR/."
cp "$ROOT_DIR/hack/test/deploy/prom-example.yaml" "$TEMP_DIR/."

if [ -n "${COLLECTOR_CONFIG_FILE_PATH}" ]; then
    echo "VAR is set to a non-empty string"
  cp "$ROOT_DIR/$COLLECTOR_CONFIG_FILE_PATH" "$TEMP_DIR/4-collector-config.yaml"
fi

pushd "$TEMP_DIR"
  sed -i '' "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" "$TEMP_DIR/6-wavefront-proxy.yaml"
  sed -i '' "s/k8s-cluster/${CONFIG_CLUSTER_NAME}/g" "$TEMP_DIR/4-collector-config.yaml"

  echo "using collector version ${CURRENT_COLLECTOR_VERSION}"
  sed -i '' "s/$DEFAULT_COLLECTOR_VERSION/$CURRENT_COLLECTOR_VERSION/g" "$TEMP_DIR/5-collector-daemonset.yaml"
  sed -i '' "s%${DEFAULT_COLLECTOR_REPO}%${CURRENT_COLLECTOR_REPO}%g" "$TEMP_DIR/5-collector-daemonset.yaml"

  echo "using proxy version ${CURRENT_PROXY_VERSION}"
  sed -i '' "s/$DEFAULT_PROXY_VERSION/$CURRENT_PROXY_VERSION/g" "$TEMP_DIR/6-wavefront-proxy.yaml"
  sed -i '' "s%${DEFAULT_PROXY_REPO}%${CURRENT_PROXY_REPO}%g" "$TEMP_DIR/6-wavefront-proxy.yaml"

  kubectl config set-context --current --namespace="$NS"
  kubectl apply -f "$TEMP_DIR/."
  kubectl config set-context --current --namespace=default
popd
