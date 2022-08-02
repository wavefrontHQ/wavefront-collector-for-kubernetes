#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-c wavefront instance name (default: 'nimba')"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-p PPS goal (required)"
  exit 1
}

function main() {
  cd "$(dirname "$0")" # cd to deploy-load-test.sh is in

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local PPS_GOAL=
  local WF_CLUSTER=nimba
  local NUMBER_OF_PROM_REPLICAS=
  local TEMP_DIR=$(mktemp -d)

  while getopts ":c:t:p:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    p)
      PPS_GOAL="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_msg_and_exit "wavefront token required"
  fi

  if [[ -z ${PPS_GOAL} ]]; then
    print_msg_and_exit "PPS goal required"
  fi

  NUMBER_OF_PROM_REPLICAS=$(awk -v pps=$PPS_GOAL 'BEGIN { printf "%3.0f\n", (pps-60)/50 }')

  if [[ ${NUMBER_OF_PROM_REPLICAS} -lt 0 ]]; then
    NUMBER_OF_PROM_REPLICAS=0
  fi

  DEPLOY_TARGETS=no ./deploy-local.sh

  green "PPS Goal: ${PPS_GOAL} Number of prom example replicas: ${NUMBER_OF_PROM_REPLICAS}"
  cp "$REPO_ROOT/hack/test/deploy/load-test-prom-example.yaml" "$TEMP_DIR/."

  kubectl delete namespace load-test || true

  pushd "$TEMP_DIR"
 		sed -i '' "s/NUMBER_OF_REPLICAS/${NUMBER_OF_PROM_REPLICAS}/g" "$TEMP_DIR/load-test-prom-example.yaml"
 		kubectl apply -f "$TEMP_DIR/."
  popd
}

main $@
