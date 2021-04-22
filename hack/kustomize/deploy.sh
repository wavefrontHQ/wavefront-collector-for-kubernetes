#! /bin/bash -e

# This script automates the deployment of the collector to a specific k8s cluster

DEFAULT_IMAGE_NAME="wavefronthq\/wavefront-kubernetes-collector"
DEFAULT_VERSION="1.3.5"
DEFAULT_FLUSH_ONCE=false

if [[ -z ${FLUSH_ONCE} ]] ; then
    FLUSH_ONCE=${DEFAULT_FLUSH_ONCE}
fi

if $FLUSH_ONCE ;
  then
    REDIRECT_TO_LOG=true
    FLUSH_INTERVAL=18
    COLLECTION_INTERVAL=10
  else
    REDIRECT_TO_LOG=false
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
    echo -e "\t-p kustomize overlay to deploy"
    exit 1
}

WF_CLUSTER=
WF_TOKEN=
VERSION=
IMAGE=
PROFILE=

while getopts "c:t:v:i:p:" opt; do
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
    p)
      PROFILE="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
  esac
done

echo "$WF_CLUSTER $VERSION $IMAGE $PROFILE"

if [[ -z ${WF_CLUSTER} || -z ${WF_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    print_usage_and_exit "wavefront instance and token required"
fi

BASE_DIR="base"
OVERLAYS_DIR="overlays"

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi
NAMESPACE_VERSION=$(echo "$VERSION" | tr . -)

if [[ -z ${IMAGE} ]] ; then
    IMAGE=${DEFAULT_IMAGE_NAME}
fi

echo "FLUSH ONCE: ${FLUSH_ONCE}"

#TODO: temp directory for intermediate files
#TODO: need to replace the kustomize template to source from temp directory

#TODO: Migrate these sed to use kustomize edit
sed "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WF_TOKEN}/g" ${BASE_DIR}/proxy.template.yaml > ${BASE_DIR}/proxy.yaml
sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" ${BASE_DIR}/kustomization.template.yaml | sed "s/YOUR_IMAGE_TAG/${VERSION}/g" | sed "s/YOUR_IMAGE_NAME/${IMAGE}/g" > ${BASE_DIR}/kustomization.yaml

sed "s/YOUR_CLUSTER_NAME/cluster-${VERSION}/g" ${OVERLAYS_DIR}/test/collector.yaml.template | sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" \
|  sed "s/FLUSH_ONCE/${FLUSH_ONCE}/g" \
|  sed "s/REDIRECT_TO_LOG/${REDIRECT_TO_LOG}/g" \
|  sed "s/FLUSH_INTERVAL/${FLUSH_INTERVAL}/g" \
|  sed "s/COLLECTION_INTERVAL/${COLLECTION_INTERVAL}/g" > ${OVERLAYS_DIR}/test/collector.yaml

kustomize build ${OVERLAYS_DIR}/test | kubectl apply -f -
