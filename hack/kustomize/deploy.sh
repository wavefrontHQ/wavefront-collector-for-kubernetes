#! /bin/bash -e

# This script automates the deployment of the collector to a specific k8s cluster

DEFAULT_IMAGE_NAME="wavefronthq\/wavefront-kubernetes-collector"

DEFAULT_VERSION="1.3.5"
USE_TEST_PROXY="${USE_TEST_PROXY:-false}"
FLUSH_ONCE="${FLUSH_ONCE:-false}"

if [ "$USE_TEST_PROXY" = true ] ;
  then
    FLUSH_ONCE=true
    FLUSH_INTERVAL=18
    COLLECTION_INTERVAL=10
  else
    FLUSH_INTERVAL=30
    COLLECTION_INTERVAL=60
fi

function print_usage_and_exit() {
    echo "Failure: $1"
    echo "Usage: $0 [flags] [options]"
    echo -e "\t-c wavefront instance name (required)"
    echo -e "\t-t wavefront token (required)"
    echo -e "\t-i collector docker image name"
    echo -e "\t-v collector docker image version"
    exit 1
}

WF_CLUSTER=
WF_TOKEN=
VERSION=
IMAGE=

while getopts "c:t:v:i:" opt; do
  case $opt in
    c)
      WF_CLUSTER="$OPTARG"
      ;;
    t)
      WF_TOKEN="$OPTARG"
      ;;
    v)
      VERSION="$OPTARG"
      ;;
    i)
      IMAGE="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
  esac
done

echo "$WF_CLUSTER $VERSION $IMAGE"

if [[ -z ${WF_CLUSTER} || -z ${WF_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    print_usage_and_exit "wavefront instance and token required"
fi


if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi
NAMESPACE_VERSION=$(echo "$VERSION" | tr . -)

if [[ -z ${IMAGE} ]] ; then
    IMAGE=${DEFAULT_IMAGE_NAME}
fi

echo "FLUSH ONCE: ${FLUSH_ONCE}"

if $USE_TEST_PROXY ; then
  sed "s/YOUR_IMAGE_TAG/${VERSION}/g" base/test-proxy.template.yaml > base/proxy.yaml
else
  sed "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WF_TOKEN}/g" base/proxy.template.yaml > base/proxy.yaml
fi

sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" base/kustomization.template.yaml | sed "s/YOUR_IMAGE_TAG/${VERSION}/g" | sed "s/YOUR_IMAGE_NAME/${IMAGE}/g" > base/kustomization.yaml

sed "s/YOUR_CLUSTER_NAME/cluster-${VERSION}/g" overlays/test/collector.yaml.template | sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" \
|  sed "s/FLUSH_ONCE/${FLUSH_ONCE}/g" \
|  sed "s/FLUSH_INTERVAL/${FLUSH_INTERVAL}/g" \
|  sed "s/COLLECTION_INTERVAL/${COLLECTION_INTERVAL}/g" > overlays/test/collector.yaml

kustomize build overlays/test | kubectl apply -f -
