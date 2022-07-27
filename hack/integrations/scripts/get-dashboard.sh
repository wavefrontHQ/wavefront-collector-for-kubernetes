#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh
excludedKeys=$(jq -r '.exclude | join(", ")' ${REPO_ROOT}/hack/integrations/key-filter.json)

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-c wavefront instance name (default: 'nimba')"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-d dashboard url"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working" # hack/integrations/scripts

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local DASHBOARD_URL=

  local WF_CLUSTER=nimba


  while getopts ":c:t:d:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    d)
      DASHBOARD_URL="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_msg_and_exit "wavefront token required"
  fi

  if [[ -z ${DASHBOARD_URL} ]]; then
    print_msg_and_exit "dashboard url required"
  fi

  curl -sX GET --header "Accept: application/json" \
    --header "Authorization: Bearer ${WAVEFRONT_TOKEN}" \
    "https://${WF_CLUSTER}.wavefront.com/api/v2/dashboard/${DASHBOARD_URL}" \
    | jq "del(.response | ${excludedKeys})"  | jq .response > ${DASHBOARD_URL}-partial-base.json

  cat ${DASHBOARD_URL}-partial-base.json  | jq "del(.sections , .parameterDetails)" > ${DASHBOARD_URL}-partial-reduced.json
  cat ${DASHBOARD_URL}-partial-base.json  |  jq '. | {"sections": .sections}' | jq --sort-keys >> ${DASHBOARD_URL}-partial-sections.json
  cat ${DASHBOARD_URL}-partial-base.json  |  jq '. | {"parameterDetails": .parameterDetails}' | jq --sort-keys >> ${DASHBOARD_URL}-partial-parameterDetails.json

  jq -s '.[0] * .[1]' \
    ${DASHBOARD_URL}-partial-reduced.json \
    ${DASHBOARD_URL}-partial-sections.json \
    > ${DASHBOARD_URL}-partial-with-sections.json

  jq -s '.[0] * .[1]' \
    ${DASHBOARD_URL}-partial-with-sections.json \
    ${DASHBOARD_URL}-partial-parameterDetails.json \
    > ${DASHBOARD_URL}.json

  rm *-partial*.json
}

main $@
