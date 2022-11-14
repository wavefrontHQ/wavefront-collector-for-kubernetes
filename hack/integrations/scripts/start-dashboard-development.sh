#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  red "Failure: $1"
  echo "Usage: $0 -t <WAVEFRONT_TOKEN> -n <NEW_DASHBOARD> -b <BRANCH_NAME_SUFFIX> -c [WF_CLUSTER] -d [DASHBOARD_TO_CLONE]"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-n new dashboard url to create (required)"
  echo -e "\t-b sets the BRANCH_NAME_SUFFIX for the branch name to be created in the integrations repo"
  echo -e "\t   with the format: 'k8po/kubernetes-<BRANCH_NAME_SUFFIX>' (required)"
  echo -e "\t-c wavefront instance name (optional, default: 'nimba')"
  echo -e "\t-d dashboard url to clone from (optional, default: 'integration-dashboard-template')"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working"

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local NEW_DASHBOARD=
  local BRANCH_NAME_SUFFIX=

  # OPTIONAL
  local DASHBOARD_TO_CLONE=integration-dashboard-template

  local WF_CLUSTER=nimba

  while getopts ":c:t:d:n:b:" opt; do
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
    b)
      BRANCH_NAME_SUFFIX="$OPTARG"
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

  if [[ -z ${BRANCH_NAME_SUFFIX} ]]; then
    print_usage_and_exit "-b sets the BRANCH_NAME_SUFFIX for the branch name to be created \
      in the integrations repo with the format: 'k8po/kubernetes-<BRANCH_NAME_SUFFIX>' (required)"
  fi

  ../scripts/get-dashboard.sh -c ${WF_CLUSTER} -t ${WAVEFRONT_TOKEN} -d ${DASHBOARD_TO_CLONE} -o ${DASHBOARD_TO_CLONE}-partial-base.json

  ../scripts/sort-dashboard.sh -i ${DASHBOARD_TO_CLONE}-partial-base.json -o ${DASHBOARD_TO_CLONE}.json

  ../scripts/clean-partials.sh # because I don't want scripts bludgeoning the '-partial-base.json'

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

  local INTEGRATIONS_REPO="$HOME/workspace/integrations"
  local BRANCH_NAME="k8po/kubernetes-${BRANCH_NAME_SUFFIX}"

  pushd_check "$INTEGRATIONS_REPO"
    git stash
    git checkout master
    git pull
    git checkout -b "${BRANCH_NAME}"
    git push --set-upstream origin "${BRANCH_NAME}"
  popd_check "$INTEGRATIONS_REPO"

  green "Dashboard uploaded at https://${WF_CLUSTER}.wavefront.com/dashboards/${NEW_DASHBOARD}"
# TODO: Command to commit the new dashboard
}

main $@
