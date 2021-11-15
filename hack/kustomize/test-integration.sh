#!/bin/bash -e
source ../deploy/k8s-utils.sh

# This script automates the functional testing of the collector

DEFAULT_VERSION=$(semver-cli inc patch "$(cat ../../release/VERSION)")

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3

K8S_ENV=$(../deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}' )

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi


NS=wavefront-collector

echo "deploying configuration for additional targets"

kubectl create namespace $NS
kubectl config set-context --current --namespace="$NS"
kubectl apply -f ../deploy/mysql-config.yaml
kubectl apply -f ../deploy/memcached-config.yaml
kubectl config set-context --current --namespace=default

echo "deploying collector $IMAGE_NAME $VERSION"

env USE_TEST_PROXY=true ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v $VERSION  -k $K8S_ENV

wait_for_cluster_ready

kubectl --namespace "$NS" port-forward deploy/wavefront-proxy 8888 &
trap 'kill $(jobs -p)' EXIT

echo "waiting for logs..."
sleep 60

DIR=$(dirname "$0")
RES=$(mktemp)

cat files/metrics.jsonl  overlays/test-$K8S_ENV/metrics/additional.jsonl  > files/combined-metrics.jsonl

while true ; do # wait until we get a good connection
  RES_CODE=$(curl --silent --output "$RES" --write-out "%{http_code}" --data-binary "@$DIR/files/combined-metrics.jsonl" "http://localhost:8888/metrics/diff")
  [[ $RES_CODE -lt 200 ]] || break
done

if [[ $RES_CODE -gt 399 ]] ; then
  red "INVALID METRICS"
  jq -r '.[]' "${RES}"
  exit 1
fi

DIFF_COUNT=$(jq "(.Missing | length)" "$RES")
EXIT_CODE=0

jq -c '.Missing[]' "$RES" | sort > missing.jsonl
jq -c '.Extra[]' "$RES" | sort > extra.jsonl

if [[ $DIFF_COUNT -gt 0 ]] ; then
  red "MISSING: $(jq "(.Missing | length)" "$RES")"
  cat missing.jsonl
  red "Extra: $(jq "(.Extra | length)" "$RES")"
  red "FAILED: METRICS OUTPUT DID NOT MATCH EXACTLY"
  echo "$RES"
  if which pbcopy > /dev/null; then
    echo "$RES" | pbcopy
  fi
  EXIT_CODE=1
else
  green "SUCCEEDED"
fi

env USE_TEST_PROXY=false ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v $VERSION  -k $K8S_ENV

exit "$EXIT_CODE"
