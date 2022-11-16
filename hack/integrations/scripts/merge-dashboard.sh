#!/bin/bash -e

SCRIPT_DIR=$(dirname $0)
REPO_ROOT=$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-c wavefront instance name (default: 'nimba')"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-d dev dashboard url (required)"
  echo -e "\t-b branch name for integration repo"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working"

  # REQUIRED
  local WF_CLUSTER=nimba
  local WAVEFRONT_TOKEN=
  local DASHBOARD_DEV_URL=
  local BRANCH_NAME=

  while getopts ":c:t:d:b:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    d)
      DASHBOARD_DEV_URL="$OPTARG"
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

  if [[ -z ${DASHBOARD_DEV_URL} ]]; then
    print_usage_and_exit "dashboard url required"
  fi

  if [[ -z ${BRANCH_NAME} ]]; then
    print_usage_and_exit "missing integrations branch"
  fi

  ../scripts/get-dashboard.sh -t ${WAVEFRONT_TOKEN} -d ${DASHBOARD_DEV_URL} -o ${DASHBOARD_DEV_URL}.json

  local INTEGRATION_DIR=${REPO_ROOT}/../integrations
  git -C "$INTEGRATION_DIR" stash
  git -C "$INTEGRATION_DIR" fetch
  git -C "$INTEGRATION_DIR" switch -C "$BRANCH_NAME"

  # Change the url field to match the integration url instead of the dev dashboard url
  local DASHBOARD_URL="integration-$(echo "${DASHBOARD_DEV_URL}" | sed 's/-dev//')"
  jq ".url = \"${DASHBOARD_URL}\"" ${DASHBOARD_DEV_URL}.json > ${DASHBOARD_URL}.json

  # Copy dashboard version from integration feature branch and increment it
  local VERSION=$(($(jq ".systemDashboardVersion" ${INTEGRATION_DIR}/kubernetes/dashboards/${DASHBOARD_URL}.json)+1))
  jq ". += {"systemDashboardVersion":${VERSION}}" ${DASHBOARD_URL}.json > "tmp" && mv "tmp" ${DASHBOARD_URL}.json

  # Do the sorting here so our systemDashboardVersion gets bumped to the top of the file
  ../scripts/sort-dashboard.sh -i ${DASHBOARD_URL}.json -o 'tmp' && mv "tmp" ${DASHBOARD_URL}.json

  cat ${DASHBOARD_URL}.json > ${INTEGRATION_DIR}/kubernetes/dashboards/${DASHBOARD_URL}.json
  echo Check your integration repo for changes.
}

main $@