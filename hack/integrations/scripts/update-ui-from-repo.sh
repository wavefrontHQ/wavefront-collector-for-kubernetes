#!/bin/bash -e

REPO_ROOT=~/workspace/wavefront-collector-for-kubernetes
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-c wavefront instance name (default: 'nimba')"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-u dev dashboard url from UI (required)"
  echo -e "\t-j json dashboard name from local integration repo"
  exit 1
}

function main() {
  cd "$(dirname "$0")/../working"

  # REQUIRED
  local WAVEFRONT_TOKEN=
  local DEV_DASHBOARD_URL=

  local LOCAL_DASHBOARD_JSON=
  local WF_CLUSTER=nimba


  while getopts ":c:t:u:j:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    u)
      DEV_DASHBOARD_URL="$OPTARG"
      ;;
    j)
      LOCAL_DASHBOARD_JSON="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_usage_and_exit "wavefront token required"
  fi

  if [[ -z ${DEV_DASHBOARD_URL} ]]; then
    print_usage_and_exit "dev dashboard url from UI required"
  fi

  if [[ -z ${LOCAL_DASHBOARD_JSON} ]]; then
    print_usage_and_exit "json dashboard name from local integration repo required"
  fi

}

main $@