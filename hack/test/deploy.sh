#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source "${REPO_ROOT}"/hack/test/deploy/k8s-utils.sh
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

function print_usage_and_exit() {
  red "Failure: $1"
  echo "Usage: $0 -c <WAVEFRONT_CLUSTER> -t <WAVEFRONT_TOKEN> -k <K8S_ENV> -v [VERSION] -n [K8S_CLUSTER_NAME] -y [COLLECTOR_YAML] -p [USE_TEST_PROXY]"
  echo "  -c wavefront instance name (required)"
  echo "  -t wavefront token (required)"
  echo "  -k K8s ENV (required)"
  echo "  -v collector docker image version (default: load from 'release/VERSION')"
  echo "  -n K8s Cluster name"
  echo "  -y collector yaml"
  echo "  -p use test proxy (default: 'false')"
  echo "  -e experimental features"
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
  # REQUIRED
  local WF_CLUSTER=
  local WAVEFRONT_TOKEN=
  local K8S_ENV=

  # OPTIONAL/DEFAULT
  local VERSION="$(cat "${REPO_ROOT}"/release/VERSION)"
  local K8S_CLUSTER_NAME=
  local COLLECTOR_YAML=
  local USE_TEST_PROXY="false"
  local EXPERIMENTAL_FEATURES=

  while getopts "c:t:v:k:n:e:y:p:" opt; do
    case $opt in
      c)  WF_CLUSTER="$OPTARG" ;;
      t)  WAVEFRONT_TOKEN="$OPTARG" ;;
      k)  K8S_ENV="$OPTARG" ;;
      v)  VERSION="$OPTARG" ;;
      n)  K8S_CLUSTER_NAME="$OPTARG" ;;
      y)  COLLECTOR_YAML="$OPTARG" ;;
      p)  USE_TEST_PROXY="$OPTARG" ;;
      e)  EXPERIMENTAL_FEATURES="$OPTARG" ;;
      \?) print_usage_and_exit "Invalid option: -$OPTARG" ;;
    esac
  done

  check_required_argument "$WF_CLUSTER" "-c <WAVEFRONT_CLUSTER> is required"
  check_required_argument "$WAVEFRONT_TOKEN" "-t <WAVEFRONT_TOKEN> is required"
  check_required_argument "$K8S_ENV" "-k <K8S_ENV> is required"

  local additional_args=""
  if [[ -n "${COLLECTOR_YAML:-}" ]]; then
    additional_args="$additional_args -y $COLLECTOR_YAML"
  fi
  if [[ -n "${EXPERIMENTAL_FEATURES:-}" ]]; then
    additional_args="$additional_args -e $EXPERIMENTAL_FEATURES"
  fi

  "${SCRIPT_DIR}"/generate.sh \
      -c "$WF_CLUSTER" \
      -t "$WAVEFRONT_TOKEN" \
      -v "$VERSION" \
      -k "$K8S_ENV" \
      -n "$K8S_CLUSTER_NAME" \
      -p "$USE_TEST_PROXY" \
      $additional_args

  kustomize build "overlays/test-$K8S_ENV" | kubectl apply -f -
}

main $@
