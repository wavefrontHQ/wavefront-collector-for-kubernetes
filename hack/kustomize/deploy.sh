#! /bin/bash -e

# This script automates the deployment of the collector to a specific k8s cluster

DEFAULT_VERSION=$(cat ../../release/VERSION)
USE_TEST_PROXY="${USE_TEST_PROXY:-false}"

function print_usage_and_exit() {
    echo "Failure: $1"
    echo "Usage: $0 [flags] [options]"
    echo -e "\t-c wavefront instance name (required)"
    echo -e "\t-t wavefront token (required)"
    echo -e "\t-v collector docker image version"
    echo -e "\t-k K8s ENV (required)"
    exit 1
}

WF_CLUSTER=
WAVEFRONT_TOKEN=
VERSION=
K8S_ENV=

while getopts ":c:t:v:d:k:" opt; do
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
echo  "deploy:48"

env USE_TEST_PROXY="$USE_TEST_PROXY" ./generate.sh -c "$WF_CLUSTER" -t "$WAVEFRONT_TOKEN" -v $VERSION  -k $K8S_ENV

echo  "deploy:52"

kustomize build overlays/test-$K8S_ENV | kubectl apply -f -

echo  "deploy:54"
