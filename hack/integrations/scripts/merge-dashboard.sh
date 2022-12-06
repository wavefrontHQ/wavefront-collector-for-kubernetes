#!/bin/bash -e

SCRIPT_DIR=$(dirname $0)
REPO_ROOT=$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 -t <WAVEFRONT_TOKEN> -s <SOURCE_DASHBOARD> -d <DEST_DASHBOARD> -b <BRANCH_NAME> -c [WF_CLUSTER]"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-s source dashboard url (required)"
  echo -e "\t-d destination dashboard url (required)"
  echo -e "\t-b branch name for integration repo (required)"
  echo -e "\t-c wavefront instance name (optional, default: 'nimba')"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working"

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local SOURCE_DASHBOARD=
  local DEST_DASHBOARD=
  local BRANCH_NAME=

  # OPTIONAL
  local WF_CLUSTER=nimba

  while getopts ":c:t:s:d:b:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    s)
      SOURCE_DASHBOARD="$OPTARG"
      ;;
    d)
      DEST_DASHBOARD="$OPTARG"
      ;;
    b)
      BRANCH_NAME="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_usage_and_exit "wavefront token required"
  fi

  if [[ -z ${SOURCE_DASHBOARD} ]]; then
    print_usage_and_exit "source dashboard url required"
  fi

  if [[ -z ${DEST_DASHBOARD} ]]; then
    print_usage_and_exit "destination dashboard url required"
  fi

  if [[ -z ${BRANCH_NAME} ]]; then
    print_usage_and_exit "integrations branch required"
  fi

  ../scripts/get-dashboard.sh -t ${WAVEFRONT_TOKEN} -d ${SOURCE_DASHBOARD} -o ${SOURCE_DASHBOARD}.json

  local INTEGRATION_DIR=${REPO_ROOT}/../integrations
  git -C "$INTEGRATION_DIR" stash
  git -C "$INTEGRATION_DIR" fetch
  git -C "$INTEGRATION_DIR" switch -C "$BRANCH_NAME"

  # Change the url field to match the integration url instead of the dev dashboard url
  jq ".url = \"${DEST_DASHBOARD}\"" ${SOURCE_DASHBOARD}.json > ${DEST_DASHBOARD}.json

  # Copy dashboard version from integration feature branch and increment it
  local VERSION=$(($(jq ".systemDashboardVersion" ${INTEGRATION_DIR}/kubernetes/dashboards/${DEST_DASHBOARD}.json 2> /dev/null)+1))
  jq ". += {"systemDashboardVersion":${VERSION}}" ${DEST_DASHBOARD}.json > "tmp" && mv "tmp" ${DEST_DASHBOARD}.json

  # Do the sorting here so our systemDashboardVersion gets bumped to the top of the file
  ${SCRIPT_DIR}/sort-dashboard.sh -i ${DEST_DASHBOARD}.json -o 'tmp' && mv "tmp" ${DEST_DASHBOARD}.json

  cat ${DEST_DASHBOARD}.json > ${INTEGRATION_DIR}/kubernetes/dashboards/${DEST_DASHBOARD}.json
  echo Check your integration repo for changes.

  green "\n===============Begin dashboard validation==============="
  ruby ${SCRIPT_DIR}/dashboards_validator.rb ${INTEGRATION_DIR}/kubernetes/dashboards/${DEST_DASHBOARD}.json
  green "================End dashboard validation================\n"

  green "Next steps: Fix any validation errors, if identified. Check your integration repo for changes and commit them."
}

main $@