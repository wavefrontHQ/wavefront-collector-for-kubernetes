#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source "${REPO_ROOT}"/hack/test/deploy/k8s-utils.sh
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
NS=wavefront-collector

function run_fake_proxy_test() {
  local METRICS_FILE_NAME=$1
  local COLLECTOR_YAML=$2
  local EXPERIMENTAL_FEATURES=$3

  local USE_TEST_PROXY="true"
  local SLEEP_TIME=70

  wait_for_cluster_resource_deleted namespace/$NS

  kubectl create namespace $NS
  kubectl config set-context --current --namespace="$NS"
  kubectl apply -f ./deploy/mysql-config.yaml
  kubectl apply -f ./deploy/memcached-config.yaml
  kubectl config set-context --current --namespace=default

  echo "deploying collector $IMAGE_NAME $VERSION"

  local additional_args=""
  if [[ -n "${COLLECTOR_YAML:-}" ]]; then
    additional_args="$additional_args -y $COLLECTOR_YAML"
  fi
  if [[ -n "${EXPERIMENTAL_FEATURES:-}" ]]; then
    additional_args="$additional_args -e $EXPERIMENTAL_FEATURES"
  fi

  "${SCRIPT_DIR}"/deploy.sh \
      -c "$WAVEFRONT_CLUSTER" \
      -t "$WAVEFRONT_TOKEN" \
      -k "$K8S_ENV" \
      -v "$VERSION" \
      -n "$K8S_CLUSTER_NAME" \
      -p "$USE_TEST_PROXY" \
      $additional_args

  wait_for_cluster_ready

  kubectl --namespace "$NS" port-forward deploy/wavefront-proxy 8888 &
  trap 'kill $(jobs -p)' EXIT

  echo "waiting for logs..."
  sleep ${SLEEP_TIME}

  RES=$(mktemp)

  if [ -f "overlays/test-$K8S_ENV/metrics/${METRICS_FILE_NAME}.jsonl" ]; then
    cat "files/${METRICS_FILE_NAME}.jsonl" overlays/test-$K8S_ENV/metrics/${METRICS_FILE_NAME}.jsonl >files/combined-metrics.jsonl
  else
    cat "files/${METRICS_FILE_NAME}.jsonl" >files/combined-metrics.jsonl
  fi

  while true; do # wait until we get a good connection
    RES_CODE=$(curl --silent --output "$RES" --write-out "%{http_code}" --data-binary "@$SCRIPT_DIR/files/combined-metrics.jsonl" "http://localhost:8888/metrics/diff")
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

  # kill port-forwarding to unbind the port and enable running the next proxy test
  kill "$(jobs -p)" || true
}

function run_real_proxy_metrics_test () {
  local USE_TEST_PROXY="false"
  local additional_args="-p $USE_TEST_PROXY"
  if [[ -n "${EXPERIMENTAL_FEATURES:-}" ]]; then
    additional_args="$additional_args -e $EXPERIMENTAL_FEATURES"
  fi

  wait_for_cluster_ready

  "${SCRIPT_DIR}"/deploy.sh \
      -c "$WAVEFRONT_CLUSTER" \
      -t "$WAVEFRONT_TOKEN" \
      -k "$K8S_ENV" \
      -v "$VERSION" \
      -n "$K8S_CLUSTER_NAME" \
      $additional_args

  "${SCRIPT_DIR}"/test-wavefront-metrics.sh -t "$WAVEFRONT_TOKEN"
  green "SUCCEEDED"
}

function print_usage_and_exit() {
  red "Failure: $1"
  echo "Usage: $0 -c <WAVEFRONT_CLUSTER> -t <WAVEFRONT_TOKEN> -v [VERSION] -r [INTEGRATION_TEST_ARGS...]"
  echo "  -c wavefront instance name (required)"
  echo "  -t wavefront token (required)"
  echo "  -v collector docker image version (default: load from 'release/VERSION')"
  echo "  -k K8s ENV"
  echo "  -n K8s Cluster name"
  echo "  -r tests to run"
  exit 1
}

function check_required_argument() {
  local required_arg=$1
  local failure_msg=$2
  if [[ -z ${required_arg} ]]; then
    print_usage_and_exit "$failure_msg"
  fi
}

function main() {
  local EXIT_CODE=0

  # REQUIRED
  local WAVEFRONT_CLUSTER=
  local WAVEFRONT_TOKEN=

  # OPTIONAL/DEFAULT
  local VERSION=$(semver-cli inc patch "$(cat "${REPO_ROOT}"/release/VERSION)")
  local K8S_ENV=$("${SCRIPT_DIR}"/deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}')
  local K8S_CLUSTER_NAME=$(whoami)-${K8S_ENV}-$(date +"%y%m%d")
  local EXPERIMENTAL_FEATURES=
  local tests_to_run=()

  while getopts ":c:t:v:k:n:r:" opt; do
    case $opt in
      c)  WAVEFRONT_CLUSTER="$OPTARG" ;;
      t)  WAVEFRONT_TOKEN="$OPTARG" ;;
      v)  VERSION="$OPTARG" ;;
      k)  K8S_ENV="$OPTARG" ;;
      n)  K8S_CLUSTER_NAME="$OPTARG" ;;
      r)  tests_to_run+=("$OPTARG") ;;
      \?) print_usage_and_exit "Invalid option: -$OPTARG" ;;
    esac
  done

  check_required_argument "$WAVEFRONT_CLUSTER" "-c <WAVEFRONT_CLUSTER> is required"
  check_required_argument "$WAVEFRONT_TOKEN" "-t <WAVEFRONT_TOKEN> is required"

  if [[ ${#tests_to_run[@]} -eq 0 ]]; then
    tests_to_run=( "default" )
  fi

  if [[ "${tests_to_run[*]}" =~ "cluster-metrics-only" ]]; then
    echo "==================== Running fake_proxy cluster-metrics-only test ===================="
    run_fake_proxy_test "cluster-metrics-only" "base/deploy/collector-deployments/5-collector-cluster-metrics-only.yaml"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "node-metrics-only" ]]; then
    echo "==================== Running fake_proxy node-metrics-only test ===================="
    run_fake_proxy_test "node-metrics-only" "base/deploy/collector-deployments/5-collector-node-metrics-only.yaml"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "combined" ]]; then
    echo "==================== Running fake_proxy combined test ===================="
    run_fake_proxy_test "all-metrics" "base/deploy/collector-deployments/5-collector-combined.yaml"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "single-deployment" ]]; then
    echo "==================== Running fake_proxy single-deployment test ===================="
    run_fake_proxy_test "all-metrics" "base/deploy/collector-deployments/5-collector-single-deployment.yaml"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "histogram-conversion" ]]; then
    echo "==================== Running fake_proxy histogram-conversion test ===================="
    run_fake_proxy_test "all-metrics" "${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml" "histogram-conversion"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "default" ]]; then
    echo "==================== Running fake_proxy default test ===================="
    run_fake_proxy_test "all-metrics" "${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml"
    ${SCRIPT_DIR}/clean-deploy.sh
  fi
  if [[ "${tests_to_run[*]}" =~ "real-proxy-metrics" ]]; then
    echo "==================== Running real-proxy-metrics test ===================="
    run_real_proxy_metrics_test
  fi

  exit "$EXIT_CODE"
}

main $@
