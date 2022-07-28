#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

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

  local DASHBOARD_DEV_URL="$(echo "${DASHBOARD_URL}" | sed 's/integration-//')-dev"

  ../scripts/get-dashboard.sh -t ${WAVEFRONT_TOKEN} -d ${DASHBOARD_URL}

  jq ".url = \"${DASHBOARD_DEV_URL}\"" ${DASHBOARD_URL}.json >  ${DASHBOARD_DEV_URL}.json

  result=$(curl -X PUT --data "$(cat "${DASHBOARD_DEV_URL}".json)" \
    --header "Content-Type: application/json" \
    --header "Authorization: Bearer ${WAVEFRONT_TOKEN}" \
    "https://${WF_CLUSTER}.wavefront.com/api/v2/dashboard/${DASHBOARD_DEV_URL}" \
    --write-out '%{http_code}' --silent --output /dev/null)
  if [[ $result -ne 200 ]]; then
    result=$(curl --silent -X POST --data "$(cat "${DASHBOARD_DEV_URL}".json)" \
      --header "Content-Type: application/json" \
      --header "Authorization: Bearer ${WAVEFRONT_TOKEN}" \
      "https://${WF_CLUSTER}.wavefront.com/api/v2/dashboard" \
      --write-out '%{http_code}' --silent --output /dev/null)
    if [[ $result -ne 200 ]]; then
      red "Uploading ${DASHBOARD_DEV_URL} dashboard failed with error code: ${result}"
    fi
  fi

}

main $@
