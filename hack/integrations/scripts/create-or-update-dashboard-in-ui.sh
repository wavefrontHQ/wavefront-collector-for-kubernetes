#!/bin/bash -e

SCRIPT_DIR=$(dirname $0)
REPO_ROOT=$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  red "Failure: $1"
  echo "Usage: $0 -t <WAVEFRONT_TOKEN> -n <NEW_DASHBOARD> -c [WF_CLUSTER] -d [DASHBOARD_TO_CLONE]"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-n new dashboard url to create (required)"
  echo -e "\t-c wavefront instance name (optional, default: 'nimba')"
  echo -e "\t-d dashboard url to clone from (optional, default: 'integration-dashboard-template')"
  exit 1
}

function main() {
  cd "${REPO_ROOT}/hack/integrations/working"

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local NEW_DASHBOARD=

  # OPTIONAL
  local DASHBOARD_TO_CLONE=integration-dashboard-template

  local WF_CLUSTER=nimba

  while getopts ":c:t:d:n:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    d)
      DASHBOARD_TO_CLONE="$OPTARG"
      ;;
    n)
      NEW_DASHBOARD="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_usage_and_exit "-t wavefront token (required)"
  fi

  if [[ -z ${NEW_DASHBOARD} ]]; then
    print_usage_and_exit "-n new dashboard url to create (required)"
  fi

  ../scripts/get-dashboard.sh -c ${WF_CLUSTER} -t ${WAVEFRONT_TOKEN} -d ${DASHBOARD_TO_CLONE} -o ${DASHBOARD_TO_CLONE}.json

  jq ".url = \"${NEW_DASHBOARD}\"" ${DASHBOARD_TO_CLONE}.json > ${NEW_DASHBOARD}.json

  local RESULT=$(curl -X PUT --data "$(cat "${NEW_DASHBOARD}".json)" \
    --header "Content-Type: application/json" \
    --header "Authorization: Bearer ${WAVEFRONT_TOKEN}" \
    "https://${WF_CLUSTER}.wavefront.com/api/v2/dashboard/${NEW_DASHBOARD}" \
    --write-out '%{http_code}' --silent --output /dev/null)
  if [[ $RESULT -ne 200 ]]; then
    RESULT=$(curl --silent -X POST --data "$(cat "${NEW_DASHBOARD}".json)" \
      --header "Content-Type: application/json" \
      --header "Authorization: Bearer ${WAVEFRONT_TOKEN}" \
      "https://${WF_CLUSTER}.wavefront.com/api/v2/dashboard" \
      --write-out '%{http_code}' --silent --output /dev/null)
    if [[ $RESULT -ne 200 ]]; then
      red "Uploading ${NEW_DASHBOARD} dashboard failed with error code: ${RESULT}"
      exit 1
    fi
  fi

  green "Dashboard uploaded at https://${WF_CLUSTER}.wavefront.com/dashboards/${NEW_DASHBOARD}"
}

main $@
