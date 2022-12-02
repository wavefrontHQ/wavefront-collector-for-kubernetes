#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source "${REPO_ROOT}"/hack/test/deploy/k8s-utils.sh
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
NS=wavefront-collector

function copy_collector_deployment_files() {
  echo "Copying collector deployment files"

  cp "${REPO_ROOT}/deploy/kubernetes/0-collector-namespace.yaml" base/deploy/0-collector-namespace.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/1-collector-cluster-role.yaml" base/deploy/1-collector-cluster-role.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/2-collector-rbac.yaml" base/deploy/2-collector-rbac.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/3-collector-service-account.yaml" base/deploy/3-collector-service-account.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml" base/deploy/collector-deployments/5-collector-daemonset.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/alternate-collector-deployments/5-collector-single-deployment.yaml" base/deploy/collector-deployments/5-collector-single-deployment.yaml
  cp "${REPO_ROOT}/deploy/kubernetes/alternate-collector-deployments/5-collector-combined.yaml" base/deploy/collector-deployments/5-collector-combined.yaml

  csplit base/deploy/collector-deployments/5-collector-combined.yaml '/^---$/' &>/dev/null
  mv xx00 base/deploy/collector-deployments/5-collector-node-metrics-only.yaml
  mv xx01 base/deploy/collector-deployments/5-collector-cluster-metrics-only.yaml

  cp "${COLLECTOR_YAML}" base/deploy/5-wavefront-collector.yaml
}

function replace_placeholders_in_template_yaml() {
  echo "Replacing placeholders in template yaml files"
  local FLUSH_INTERVAL=30
  local COLLECTION_INTERVAL=60

  if [[ "${USE_TEST_PROXY}" = "true" ]]; then
    FLUSH_INTERVAL=18
    COLLECTION_INTERVAL=7
    cp base/test-proxy.yaml base/deploy/6-wavefront-proxy.yaml
  else
    sed -e "s/YOUR_CLUSTER/${WF_CLUSTER}/g" \
        -e "s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" \
        base/proxy.template.yaml > base/deploy/6-wavefront-proxy.yaml
  fi

  sed "s/YOUR_IMAGE_TAG/${VERSION}/g" base/kustomization.template.yaml > base/kustomization.yaml

  sed -e "s/NAMESPACE/${NS}/g" \
      -e "s/YOUR_CLUSTER_NAME/${K8S_CLUSTER_NAME}/g" \
      -e "s/YOUR_EXPERIMENTAL_FEATURES/${EXPERIMENTAL_FEATURES}/g" \
      -e "s/FLUSH_INTERVAL/${FLUSH_INTERVAL}/g" \
      -e "s/COLLECTION_INTERVAL/${COLLECTION_INTERVAL}/g" \
      base/collector-config.template.yaml > base/deploy/4-collector-config.yaml
}

function print_usage_and_exit() {
  red "Failure: $1"
  echo "Usage: $0 -c <WAVEFRONT_CLUSTER> -t <WAVEFRONT_TOKEN> -v [VERSION] -k [K8S_ENV] -n [K8S_CLUSTER_NAME] -y [COLLECTOR_YAML] -p [USE_TEST_PROXY]"
  echo "  -c wavefront instance name (required)"
  echo "  -t wavefront token (required)"
  echo "  -v collector docker image version (default: load from 'release/VERSION')"
  echo "  -k K8s ENV"
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
  cd "${SCRIPT_DIR}" # hack/test

  # REQUIRED
  local WF_CLUSTER=
  local WAVEFRONT_TOKEN=

  # OPTIONAL/DEFAULT
  local VERSION="$(cat "${REPO_ROOT}"/release/VERSION)"
  local K8S_ENV=$("${SCRIPT_DIR}"/deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}' )
  local K8S_CLUSTER_NAME=$(whoami)-${K8S_ENV}-$(date +"%y%m%d")
  local COLLECTOR_YAML="${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml"
  local USE_TEST_PROXY="false"
  local EXPERIMENTAL_FEATURES=

  while getopts "c:t:v:k:n:e:y:p:" opt; do
    case $opt in
      c)  WF_CLUSTER="$OPTARG" ;;
      t)  WAVEFRONT_TOKEN="$OPTARG" ;;
      v)  VERSION="$OPTARG" ;;
      k)  K8S_ENV="$OPTARG" ;;
      n)  K8S_CLUSTER_NAME="$OPTARG" ;;
      y)  COLLECTOR_YAML="$OPTARG" ;;
      p)  USE_TEST_PROXY="$OPTARG" ;;
      e)  EXPERIMENTAL_FEATURES="$OPTARG" ;;
      \?) print_usage_and_exit "Invalid option: -$OPTARG" ;;
    esac
  done

  check_required_argument "$WF_CLUSTER" "-c <WAVEFRONT_CLUSTER> is required"
  check_required_argument "$WAVEFRONT_TOKEN" "-t <WAVEFRONT_TOKEN> is required"

  echo "Generating deployment files for wavefront collector"
  copy_collector_deployment_files
  replace_placeholders_in_template_yaml

  if [[ "${USE_TEST_PROXY}" = "true" ]] && [[ "${VERSION}" != "fake" ]]; then
    echo "IMAGE TAG: ${VERSION}"
    green "WF Cluster Name: ${K8S_CLUSTER_NAME}"
  fi
}

main $@
