#!/bin/bash -e
source ./deploy/k8s-utils.sh

# This script automates the functional testing of the collector

DEFAULT_VERSION=$(semver-cli inc patch "$(cat ../../release/VERSION)")

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3
INTEGRATION_TEST_TYPE=$4

K8S_ENV=$(./deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}' )

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

METRICS_FILE_NAME=all-metrics
COLLECTOR_YAML=
SLEEP_TIME=70

if [[ "${INTEGRATION_TEST_TYPE}" == "cluster-metrics-only" ]]; then
  METRICS_FILE_NAME="cluster-metrics-only"
  COLLECTOR_YAML="base/deploy/collector-deployments/5-collector-cluster-metrics-only.yaml"
fi
if [[ "${INTEGRATION_TEST_TYPE}" == "node-metrics-only" ]]; then
  METRICS_FILE_NAME="node-metrics-only"
  COLLECTOR_YAML="base/deploy/collector-deployments/5-collector-node-metrics-only.yaml"
fi
if [[ "${INTEGRATION_TEST_TYPE}" == "combined" ]]; then
  COLLECTOR_YAML="base/deploy/collector-deployments/5-collector-combined.yaml"
fi
if [[ "${INTEGRATION_TEST_TYPE}" == "single-deployment" ]]; then
  METRICS_FILE_NAME="legacy-deployment"
  COLLECTOR_YAML="base/deploy/collector-deployments/5-collector-single-deployment.yaml"
fi

NS=wavefront-collector

echo "deploying configuration for additional targets"

kubectl create namespace $NS
kubectl config set-context --current --namespace="$NS"
kubectl apply -f ./deploy/mysql-config.yaml
kubectl apply -f ./deploy/memcached-config.yaml
kubectl config set-context --current --namespace=default

echo "deploying collector $IMAGE_NAME $VERSION"

env USE_TEST_PROXY=true ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v "$VERSION"  -k "$K8S_ENV" -y "$COLLECTOR_YAML"

wait_for_cluster_ready

kubectl --namespace "$NS" port-forward deploy/wavefront-proxy 8888 &
trap 'kill $(jobs -p)' EXIT

echo "waiting for logs..."
sleep ${SLEEP_TIME}

DIR=$(dirname "$0")
RES=$(mktemp)

cat "files/${METRICS_FILE_NAME}.jsonl"  overlays/test-$K8S_ENV/metrics/${METRICS_FILE_NAME}.jsonl  > files/combined-metrics.jsonl

while true ; do # wait until we get a good connection
  RES_CODE=$(curl --silent --output "$RES" --write-out "%{http_code}" --data-binary "@$DIR/files/combined-metrics.jsonl" "http://localhost:8888/metrics/diff")
  [[ $RES_CODE -lt 200 ]] || break
done

if [[ $RES_CODE -gt 399 ]] ; then
  red "INVALID METRICS"
  jq -r '.[]' "${RES}"
  exit 1
fi

DIFF_COUNT=$(jq "(.Missing | length) + (.Unwanted | length)" "$RES")
EXIT_CODE=0

jq -c '.Missing[]' "$RES" | sort > missing.jsonl
jq -c '.Extra[]' "$RES" | sort > extra.jsonl
jq -c '.Unwanted[]' "$RES" | sort > unwanted.jsonl

echo "$RES"
if [[ $DIFF_COUNT -gt 0 ]] ; then
  red "Missing: $(jq "(.Missing | length)" "$RES")"
  if [[ $(jq "(.Missing | length)" "$RES") -le 10 ]]; then
    cat missing.jsonl
  fi
  red "Unwanted: $(jq "(.Unwanted | length)" "$RES")"
  if [[ $(jq "(.Unwanted | length)" "$RES") -le 10 ]]; then
    cat unwanted.jsonl
  fi
  red "Extra: $(jq "(.Extra | length)" "$RES")"
  red "FAILED: METRICS OUTPUT DID NOT MATCH"
  if which pbcopy > /dev/null; then
    echo "$RES" | pbcopy
  fi
  EXIT_CODE=1
else
  green "SUCCEEDED"
fi

env USE_TEST_PROXY=false ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v "$VERSION"  -k "$K8S_ENV" -y "$COLLECTOR_YAML"

exit "$EXIT_CODE"
