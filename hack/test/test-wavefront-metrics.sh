#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function curl_query_to_wf_dashboard() {
  local query=$1
  # NOTE: any output inside this function is concatenated and used as the return value;
  # otherwise we would love to put a log such as this in here to give us more information:
  # echo "=============== Querying '$WF_CLUSTER' for query '${query}'"
  curl -X GET --header "Accept: application/json" \
    --header "Authorization: Bearer $WAVEFRONT_TOKEN" \
    "https://$WF_CLUSTER.wavefront.com/api/v2/chart/api?q=${query}&queryType=WQL&s=$AFTER_UNIX_TS&g=s&view=METRIC&sorted=false&cached=true&useRawQK=false" |
    jq '.timeseries[0].data[0][1]'
}

function wait_for_query_match_exact() {
  local query_match_exact=$1
  local expected=$2
  local actual
  local loop_count=0
  while [[ $loop_count -lt $MAX_QUERY_TIMES ]]; do
    loop_count=$((loop_count + 1))
    echo "===============BEGIN checking wavefront dashboard metrics for '$query_match_exact' - attempt $loop_count/$MAX_QUERY_TIMES"
    actual=$(curl_query_to_wf_dashboard "${query_match_exact}")
    echo "Actual is: '$actual'"
    echo "Expected is '${expected}'"
    echo "===============END checking wavefront dashboard metrics for $query_match_exact"

    if echo "$actual $expected" | awk '{exit ($1 > $2 || $1 < $2)}'; then
        return 0
    fi

    sleep $CURL_WAIT
  done

  return 1
}

function wait_for_query_non_zero() {
  local query_non_zero=$1
  local actual=0
  local loop_count=0
  while [[ $actual == null || $actual == 0 ]] && [[ $loop_count -lt $MAX_QUERY_TIMES ]]; do
    loop_count=$((loop_count + 1))

    echo "===============BEGIN checking wavefront dashboard metrics for $query_non_zero - attempt $loop_count/$MAX_QUERY_TIMES"
    actual=$(curl_query_to_wf_dashboard "${query_non_zero}")
    echo "Actual is: '$actual'"
    echo "Expected non zero"
    echo "===============END checking wavefront dashboard metrics for $query_non_zero"

    sleep $CURL_WAIT
  done

  if [[ $actual == null || $actual == 0 ]]; then
    return 1
  fi
}

function print_usage_and_exit() {
  echo "Failure: $1"
  echo "Usage: $0 [flags] [options]"
  echo -e "\t-c wavefront instance name (default: 'nimba')"
  echo -e "\t-t wavefront token (required)"
  echo -e "\t-n config cluster name for metric grouping (default: \$(whoami)-<default version from file>-release-test)"
  echo -e "\t-v collector docker image version (default: load from 'release/VERSION')"
  exit 1
}

function exit_on_fail() {
  $@ # run all arguments as a command
  local exit_code=$?
  if [[ $exit_code != 0 ]]; then
    echo "Command '$@' exited with exit code '$exit_code'"
    exit $exit_code
  fi
}

function main() {
  cd "$(dirname "$0")" # hack/test

  local AFTER_UNIX_TS="$(date '+%s')000"
  local MAX_QUERY_TIMES=20
  local CURL_WAIT=15

  # REQUIRED
  local WAVEFRONT_TOKEN=

  local WF_CLUSTER=nimba
  local VERSION="$(cat ../../release/VERSION)"
  local K8S_ENV=$(./deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}' )
  local CONFIG_CLUSTER_NAME=$(whoami)-${K8S_ENV}-$(date +"%y%m%d")

  while getopts ":c:t:n:v:" opt; do
    case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    n)
      CONFIG_CLUSTER_NAME="$OPTARG"
      ;;
    v)
      VERSION="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
    esac
  done

  if [[ -z ${WAVEFRONT_TOKEN} ]]; then
    print_msg_and_exit "wavefront token required"
  fi

  local VERSION_IN_DECIMAL="${VERSION%.*}"
  local VERSION_IN_DECIMAL+="$(echo "${VERSION}" | cut -d '.' -f3)"
  local VERSION_IN_DECIMAL="$(echo "${VERSION_IN_DECIMAL}" | sed 's/0$//')"

  wait_for_cluster_ready

  exit_on_fail wait_for_query_match_exact "ts(kubernetes.collector.version%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22%20AND%20installation_method%3D%22manual%22)" "${VERSION_IN_DECIMAL}"
  exit_on_fail wait_for_query_non_zero "ts(kubernetes.cluster.pod.count%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)"
  exit_on_fail wait_for_query_non_zero "ts(mysql.connections%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)"

  local PROM_EXAMPLE_EXPECTED_COUNT="3"
  exit_on_fail wait_for_query_match_exact "ts(prom-example.schedule.activity.decision.counter%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)" "${PROM_EXAMPLE_EXPECTED_COUNT}"
}

main $@
