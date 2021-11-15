#!/usr/bin/env bash
set -e

DEFAULT_VERSION="1.7.4"
USE_TEST_PROXY="${USE_TEST_PROXY:-false}"

if [ "$USE_TEST_PROXY" = true ] ;
  then
    FLUSH_INTERVAL=18
    COLLECTION_INTERVAL=7
  else
    FLUSH_INTERVAL=30
    COLLECTION_INTERVAL=60
fi

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
K8S_ENV=gke

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
    k)
      K8S_ENV="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
  esac
done

echo "$WF_CLUSTER $VERSION $IMAGE"

if [[ -z ${WF_CLUSTER} || -z ${WAVEFRONT_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    print_usage_and_exit "wavefront instance and token required"
fi


if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

NAMESPACE_NAME=wavefront-collector


if $USE_TEST_PROXY ; then
  cp base/test-proxy.yaml base/proxy.yaml
else
  sed "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" base/proxy.template.yaml > base/proxy.yaml
fi

 sed "s/YOUR_IMAGE_TAG/${VERSION}/g" base/kustomization.template.yaml  > base/kustomization.yaml

sed "s/YOUR_CLUSTER_NAME/$(whoami)-${K8S_ENV}-${VERSION}/g"  base/collector.template.yaml  |
  sed "s/NAMESPACE/${NAMESPACE_NAME}/g" |
  sed "s/FLUSH_INTERVAL/${FLUSH_INTERVAL}/g" |
  sed  "s/COLLECTION_INTERVAL/${COLLECTION_INTERVAL}/g" > base/collector.yaml

