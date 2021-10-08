#!/bin/bash -e

cd "$(dirname "$0")" # hack/kustomize
source ../deploy/k8s-utils.sh

MAX_QUERY_TIMES=10
CURL_WAIT=10
EXIT_CODE=0

if [[ -z ${WF_CLUSTER} ]] ; then
    WF_CLUSTER=nimba
fi

if [[ -z ${WAVEFRONT_TOKEN} ]] ; then
    print_msg_and_exit "wavefront token required"
fi

if [[ -z ${CONFIG_CLUSTER_NAME} ]] ; then
    CONFIG_CLUSTER_NAME=$(whoami)-${VERSION}-release-test
fi

VERSION_FROM_FILE="$(cat ../../release/VERSION)"
if [[ -z ${VERSION} ]] ; then
    VERSION=${VERSION_FROM_FILE}
fi

AFTER_UNIX_TS="$(date '+%s')000"

wait_for_cluster_ready

function curl_query_to_wf_dashboard() {
  local query=$1
  actual=$(curl -X GET --header "Accept: application/json" \
   --header "Authorization: Bearer $WAVEFRONT_TOKEN" \
    "https://$WF_CLUSTER.wavefront.com/api/v2/chart/api?q=${query}&queryType=WQL&s=$AFTER_UNIX_TS&g=s&view=METRIC&sorted=false&cached=true&useRawQK=false" \
    | jq '.timeseries[0].data[0][1]')
}

function wait_for_query_match_exact() {
  local query_match_exact=$1
  local expected=$2
  local actual
  local loop_count=0
  while [[ $actual != "${expected}" ]] && [[ $loop_count -lt $MAX_QUERY_TIMES ]]; do
    loop_count=$((loop_count+1))
    echo "===============BEGIN checking wavefront dashboard metrics for $query_match_exact"
    echo "Trying query for $loop_count/$MAX_QUERY_TIMES times"
    curl_query_to_wf_dashboard "${query_match_exact}"
    echo "Actual is: '$actual'"
    echo "Expected is '${expected}'"
    echo "===============END checking wavefront dashboard metrics for $query_match_exact"

    sleep $CURL_WAIT
  done

  if [[ $actual != "${expected}" ]] ; then
    EXIT_CODE=1
  fi
}

function wait_for_query_non_zero() {
  local query_non_zero=$1
  local actual=0
  local loop_count=0
  while [[ $actual == null || $actual == 0 ]] && [[ $loop_count -lt $MAX_QUERY_TIMES ]]; do
    loop_count=$((loop_count+1))

    echo "===============BEGIN checking wavefront dashboard stuff for $query_non_zero"
    echo "Trying query for $loop_count/$MAX_QUERY_TIMES times"
    curl_query_to_wf_dashboard "${query_non_zero}"
    echo "Actual is: '$actual'"
    echo "Expected non zero"
    echo "===============END checking wavefront dashboard stuff for $query_non_zero"

    sleep $CURL_WAIT
  done

  if [[ $actual == null || $actual == 0 ]] ; then
    EXIT_CODE=1
  fi
}

VERSION_IN_DECIMAL="${VERSION%.*}"
VERSION_IN_DECIMAL+="$(echo "${VERSION}" | cut -d '.' -f3)"
PROM_EXAMPLE_EXPECTED_COUNT="3"

wait_for_query_match_exact "ts(kubernetes.collector.version%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22%20AND%20installation_method%3D%22manual%22)" "${VERSION_IN_DECIMAL}"
wait_for_query_non_zero "ts(kubernetes.cluster.pod.count%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)"
wait_for_query_non_zero "ts(mysql.connections%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)"
wait_for_query_match_exact "ts(prom-example.schedule.activity.decision.counter%2C%20cluster%3D%22${CONFIG_CLUSTER_NAME}%22)" "${PROM_EXAMPLE_EXPECTED_COUNT}"

echo "Exit with code '$EXIT_CODE'"
exit "$EXIT_CODE"