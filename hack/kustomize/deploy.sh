#! /bin/bash -e

# This script automates the deployment of the collector to a specific k8s cluster
DEFAULT_DOCKER_HOST="wavefronthq"

DEFAULT_VERSION="1.6.0"
USE_TEST_PROXY="${USE_TEST_PROXY:-false}"
FLUSH_ONCE="${USE_TEST_PROXY:-false}"

function print_usage_and_exit() {
    echo "Failure: $1"
    echo "Usage: $0 [flags] [options]"
    echo -e "\t-c wavefront instance name (required)"
    echo -e "\t-d docker host (required)"
    echo -e "\t-t wavefront token (required)"
    echo -e "\t-v collector docker image version"
    echo -e "\t-k K8s ENV (required)"
    exit 1
}

WF_CLUSTER=
WAVEFRONT_TOKEN=
VERSION=
DOCKER_HOST=
K8S_ENV=

while getopts "c:t:v:d:k:" opt; do
  case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WAVEFRONT_TOKEN="$OPTARG"
      ;;
    v)
      VERSION="$OPTARG"
      ;;
    d)
      DOCKER_HOST="$OPTARG"
      ;;
    k)
      K8S_ENV="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
  esac
done

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

NAMESPACE_NAME=wavefront-collector

if [[ -z ${DOCKER_HOST} ]] ; then
    DOCKER_HOST=${DEFAULT_DOCKER_HOST}
fi

env USE_TEST_PROXY="$USE_TEST_PROXY" ./generate.sh -c "$WF_CLUSTER" -t "$WAVEFRONT_TOKEN" -v $VERSION -d $DOCKER_HOST -k $K8S_ENV

# if the collector doesn't get redeployed, then the timing of picking up the
# FLUSH_ONCE config change creates inconsistent outputs

# also, if we uploaded a new collector image and didn't change the daemonset,
# will it get picked up?
if [[ $FLUSH_ONCE == "true" ]]; then
  kubectl delete namespace $NAMESPACE_NAME || true
fi

kustomize build overlays/test-$K8S_ENV | kubectl apply -f -
