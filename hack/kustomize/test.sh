#!/bin/bash -e
source ../deploy/k8s-utils.sh

# This script automates the functional testing of the collector

function green {
    echo -e $'\e[32m'$1$'\e[0m'
}

function red {
    echo -e $'\e[31m'$1$'\e[0m'
}

function print_msg_and_exit() {
    red "$1"
    exit 1
}

DEFAULT_VERSION="1.3.3"
DEFAULT_IMAGE_NAME="wavefronthq\/wavefront-kubernetes-collector"

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3
IMAGE_NAME=$4

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

if [[ -z ${IMAGE_NAME} ]] ; then
    IMAGE_NAME=${DEFAULT_IMAGE_NAME}
fi

echo "deploying collector $IMAGE_NAME $VERSION"

env USE_TEST_PROXY=true ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v $VERSION -i $IMAGE_NAME

NAMESPACE_VERSION=$(echo "${VERSION}" | tr . -)
NS=${NAMESPACE_VERSION}-wavefront-collector

echo "deploying configuration for additional targets"

kubectl config set-context --current --namespace="$NS"
kubectl apply -f ../deploy/mysql-config.yaml
kubectl apply -f ../deploy/memcached-config.yaml
kubectl config set-context --current --namespace=default

wait_for_cluster_ready

kubectl --namespace "$NS" port-forward deploy/wavefront-proxy 8888 &
trap 'kill $(jobs -p)' EXIT

echo "waiting for logs..."
sleep 30

DIR=$(dirname "$0")
RES=$(mktemp)
while true ; do # wait until we get a good connection
  RES_CODE=$(curl --silent --output "$RES" --write-out "%{http_code}" --data-binary "@$DIR/files/metrics.jsonl" "http://localhost:8888/metrics/diff")
  [[ $RES_CODE -eq 0 ]] || break
done

if [[ $RES_CODE -gt 399 ]] ; then
  red "INVALID METRICS"
  jq -r '.[]' "${RES}"
  exit 1
fi

DIFF_COUNT=$(jq "(.Missing | length) + (.Extra | length)" "$RES")
EXIT_CODE=0
if [[ $DIFF_COUNT -gt 0 ]] ; then
  jq "." "$RES"
  red "FAILED: METRICS OUTPUT DID NOT MATCH EXACTLY"
  EXIT_CODE=1
else
  green "SUCCEEDED"
fi

env USE_TEST_PROXY=false ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v $VERSION -i $IMAGE_NAME

exit "$EXIT_CODE"
