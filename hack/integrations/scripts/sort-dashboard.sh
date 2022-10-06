#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: TODO"
#  echo "Usage: $0 [flags] [options]"
#  echo -e "\t-c wavefront instance name (default: 'nimba')"
#  echo -e "\t-t wavefront token (required)"
#  echo -e "\t-d dashboard url (required)"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working"

  # REQUIRED
  local DASHBOARD_INPUT_FILE=
  local DASHBOARD_OUTPUT_FILE=

  while getopts ":i:o:" opt; do
    case $opt in
    i)
      DASHBOARD_INPUT_FILE="$OPTARG"
      ;;
    o)
      DASHBOARD_OUTPUT_FILE="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${DASHBOARD_INPUT_FILE} ]]; then
    print_msg_and_exit "dashboard input file required"
  fi

  if [[ -z ${DASHBOARD_OUTPUT_FILE} ]]; then
    print_msg_and_exit "dashboard output file required"
  fi

  cat ${DASHBOARD_INPUT_FILE} | jq "del(.sections , .parameterDetails)" > ${DASHBOARD_INPUT_FILE}.partial-reduced
  cat ${DASHBOARD_INPUT_FILE} | jq '. | {"sections": .sections}' | jq --sort-keys >> ${DASHBOARD_INPUT_FILE}.partial-sections
  cat ${DASHBOARD_INPUT_FILE} | jq '. | {"parameterDetails": .parameterDetails}' | jq --sort-keys >> ${DASHBOARD_INPUT_FILE}.partial-parameterDetails

  jq -s '.[0] * .[1]' \
    ${DASHBOARD_INPUT_FILE}.partial-reduced \
    ${DASHBOARD_INPUT_FILE}.partial-sections \
    > ${DASHBOARD_INPUT_FILE}.partial-with-sections

  jq -s '.[0] * .[1]' \
    ${DASHBOARD_INPUT_FILE}.partial-with-sections \
    ${DASHBOARD_INPUT_FILE}.partial-parameterDetails \
    > ${DASHBOARD_OUTPUT_FILE}
}

main $@
