#!/bin/bash -e

cd "$(dirname $0)"
source ../deploy/k8s-utils.sh

# This script automates the functional testing of the collector
VERSION_FROM_FILE="$(cat ../../release/VERSION)"
DEFAULT_VERSION=$(semver-cli inc patch $VERSION_FROM_FILE)
DEFAULT_DOCKER_HOST="wavefronthq"

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3
DOCKER_HOST=$4

CURL_WAIT=10

K8S_ENV=$(../deploy/get-k8s-cluster-env.sh | awk '{print tolower($0)}' )

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

if [[ -z ${DOCKER_HOST} ]] ; then
    DOCKER_HOST=${DEFAULT_DOCKER_HOST}
fi


NS=wavefront-collector

echo "deploying configuration for additional targets"

kubectl create namespace $NS
kubectl config set-context --current --namespace="$NS"
kubectl apply -f ../deploy/mysql-config.yaml
kubectl apply -f ../deploy/memcached-config.yaml
kubectl config set-context --current --namespace=default

echo "deploying collector $IMAGE_NAME $VERSION"

env USE_TEST_PROXY=false ./deploy.sh -c "$WAVEFRONT_CLUSTER" -t "$API_TOKEN" -v $VERSION -d $DOCKER_HOST -k $K8S_ENV

wait_for_cluster_ready

AFTER_UNIX_TS="$(date '+%s')000"

CLUSTER_NAME=$(whoami)-${K8S_ENV}-${VERSION}
# TODO: generate a unique cluster name or label upon each iteration to ensure entirely new metrics
# example installation_method="e2e-manual-run-<random-string>"

function waitForQueryMatchExact() {
  local query=$1
  local expected=$2
  local actual
  while [[ $actual != $expected ]]; do
    echo "-@-@-@-@-@-BEGIN checking wavefront dashboard stuff for $query-@-@-@-@-@-"
    actual=$(curl -X GET --header "Accept: application/json" \
     --header "Authorization: Bearer $API_TOKEN" \
     "https://$WAVEFRONT_CLUSTER.wavefront.com/api/v2/chart/api?q=${query}&queryType=WQL&s=$AFTER_UNIX_TS&g=s&view=METRIC&sorted=false&cached=true&useRawQK=false" \
     | jq '.timeseries[0].data[0][1]')

    echo "Actual is: '$actual'"
    echo "Expected is '$expected'"
    echo "-@-@-@-@-@-END checking wavefront dashboard stuff for $query-@-@-@-@-@-"

    sleep $CURL_WAIT
  done
}

# TODO: Yet to be verified for implementation
function waitForQueryExists() {
  local query=$1
  local actual
  while [[ $actual != 0 ]]; do
    echo "-@-@-@-@-@-BEGIN checking wavefront dashboard stuff for $query-@-@-@-@-@-"
    actual=$(curl -X GET --header "Accept: application/json" \
     --header "Authorization: Bearer $API_TOKEN" \
     "https://$WAVEFRONT_CLUSTER.wavefront.com/api/v2/chart/api?q=${query}&queryType=WQL&s=$AFTER_UNIX_TS&g=s&view=METRIC&sorted=false&cached=true&useRawQK=false" \
     | jq '.warnings')

    echo "Actual is: '$actual'"
    echo "-@-@-@-@-@-END checking wavefront dashboard stuff for $query-@-@-@-@-@-"

    sleep $CURL_WAIT
  done
}

VERSION_IN_DECIMAL="${VERSION_FROM_FILE%.*}"
VERSION_IN_DECIMAL+="$(echo ${VERSION_FROM_FILE} | cut -d '.' -f3)"
PROM_EXAMPLE_EXPECTED_COUNT="3"

# TODO: At this point it is an endless loop of querying. Need to add the actual checking and quitting the loop.
waitForQueryMatchExact "ts(kubernetes.collector.version%2C%20cluster%3D%22$CLUSTER_NAME%22%20AND%20installation_method%3D%22manual%22)" "${VERSION_IN_DECIMAL}"
#waitForQueryExists "ts(kubernetes.cluster.pod.count%2C%20cluster%3D%22$CLUSTER_NAME%22)" # TODO: Yet to be query matched
#waitForQueryExists "ts(mysql.connections%2C%20cluster%3D%22$CLUSTER_NAME%22)" # TODO: Yet to be query matched
#waitForQueryExists "ts(mysql.connectionstest%2C%20cluster%3D%22$CLUSTER_NAME%22)" # TODO: Using this to test a failure scenario error message.
waitForQueryMatchExact "ts(prom-example.schedule.activity.decision.counter%2C%20cluster%3D%22$CLUSTER_NAME%22)" "${PROM_EXAMPLE_EXPECTED_COUNT}"

#exit "$EXIT_CODE" # TODO: Not sure if we need any manipulation with $EXIT_CODE variable here, similar to test-integration.sh
