#!/bin/bash -e
source ./deploy/k8s-utils.sh
# This script automates the functional testing of the collector

function run_fake_proxy_test() {
  local METRICS_FILE_NAME=$1
  local COLLECTOR_YAML=$2
  local EXPERIMENTAL_FEATURES=$3

  echo "deploying configuration for additional targets"
  echo "EXPERIMENTAL_FEATURES is $EXPERIMENTAL_FEATURES"
  wait_for_cluster_resource_deleted namespace/$NS

  kubectl create namespace $NS
  kubectl config set-context --current --namespace="$NS"
  kubectl apply -f ./deploy/mysql-config.yaml
  kubectl apply -f ./deploy/memcached-config.yaml
  kubectl config set-context --current --namespace=default

  echo "deploying collector $IMAGE_NAME $VERSION"

  env USE_TEST_PROXY=true ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$WAVEFRONT_TOKEN" -v "$VERSION" -k "$K8S_ENV" -n "$WF_CLUSTER_NAME" -e "$EXPERIMENTAL_FEATURES" -y "$COLLECTOR_YAML"

  wait_for_cluster_ready

  kubectl --namespace "$NS" port-forward deploy/wavefront-proxy 8888 &
  trap 'kill $(jobs -p)' EXIT

  echo "waiting for logs..."
  sleep ${SLEEP_TIME}

  DIR=$(dirname "$0")
  RES=$(mktemp)

  if [ -f "overlays/test-$K8S_ENV/metrics/${METRICS_FILE_NAME}.jsonl" ]; then
    cat "files/${METRICS_FILE_NAME}.jsonl" overlays/test-$K8S_ENV/metrics/${METRICS_FILE_NAME}.jsonl >files/combined-metrics.jsonl
  else
    cat "files/${METRICS_FILE_NAME}.jsonl" >files/combined-metrics.jsonl
  fi

  while true; do # wait until we get a good connection
    RES_CODE=$(curl --silent --output "$RES" --write-out "%{http_code}" --data-binary "@$DIR/files/combined-metrics.jsonl" "http://localhost:8888/metrics/diff")
    [[ $RES_CODE -lt 200 ]] || break
  done

  if [[ $RES_CODE -gt 399 ]]; then
    red "INVALID METRICS"
    jq -r '.[]' "${RES}"
    exit 1
  fi

  DIFF_COUNT=$(jq "(.Missing | length) + (.Unwanted | length)" "$RES")
  EXIT_CODE=0

  jq -c '.Missing[]' "$RES" | sort >missing.jsonl
  jq -c '.Extra[]' "$RES" | sort >extra.jsonl
  jq -c '.Unwanted[]' "$RES" | sort >unwanted.jsonl

  echo "$RES"
  if [[ $DIFF_COUNT -gt 0 ]]; then
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
    if which pbcopy >/dev/null; then
      echo "$RES" | pbcopy
    fi
    EXIT_CODE=1
  else
    green "SUCCEEDED"
  fi
}

function run_real_proxy_metrics_test () {
  env USE_TEST_PROXY=false ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$WAVEFRONT_TOKEN" -v "$VERSION" -k "$K8S_ENV" -n "$WF_CLUSTER_NAME" -e "$EXPERIMENTAL_FEATURES" -y "$COLLECTOR_YAML"
  ./test-wavefront-metrics.sh -t $WAVEFRONT_TOKEN
  green "SUCCEEDED"
}

function main() {
  local WAVEFRONT_CLUSTER=
  local WAVEFRONT_TOKEN=
  local VERSION=$(semver-cli inc patch "$(cat ../../release/VERSION)")
  local tests_to_run=()

  local K8S_ENV=$(./deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}')

  local NS=wavefront-collector
  local SLEEP_TIME=70
  local WF_CLUSTER_NAME=$(whoami)-${K8S_ENV}-$(date +"%y%m%d")
  local EXPERIMENTAL_FEATURES=
  local EXIT_CODE=0

  while getopts ":c:t:v:r:" opt; do
    case $opt in
    c)
      WAVEFRONT_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    v)
      VERSION="$OPTARG"
      ;;
    r)
      tests_to_run+=("$OPTARG")
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done
  echo "tests_to_run is $tests_to_run"
  if [[ ${#tests_to_run[@]} -eq 0 ]]; then
    tests_to_run=(
      "default"
    )
  fi

  if [[ "${tests_to_run[*]}" =~ "cluster-metrics-only" ]]; then
    run_fake_proxy_test "cluster-metrics-only" "base/deploy/collector-deployments/5-collector-cluster-metrics-only.yaml"
  fi
  if [[ "${tests_to_run[*]}" =~ "node-metrics-only" ]]; then
    run_fake_proxy_test "node-metrics-only" "base/deploy/collector-deployments/5-collector-node-metrics-only.yaml"
  fi
  if [[ "${tests_to_run[*]}" =~ "combined" ]]; then
    run_fake_proxy_test "all-metrics" "base/deploy/collector-deployments/5-collector-combined.yaml"
  fi
  if [[ "${tests_to_run[*]}" =~ "single-deployment" ]]; then
    run_fake_proxy_test "all-metrics" "base/deploy/collector-deployments/5-collector-single-deployment.yaml"
  fi
  if [[ "${tests_to_run[*]}" =~ "histogram-conversion" ]]; then
    run_fake_proxy_test "all-metrics" "../../deploy/kubernetes/5-collector-daemonset.yaml" "histogram-conversion"
  fi
  if [[ "${tests_to_run[*]}" =~ "real-proxy-metrics" ]]; then
    run_real_proxy_metrics_test
  fi
  if [[ "${tests_to_run[*]}" =~ "default" ]]; then
    run_fake_proxy_test "all-metrics" "../../deploy/kubernetes/5-collector-daemonset.yaml"
  fi

  exit "$EXIT_CODE"
}

main $@
