#! /bin/bash

# This script automates the deployment of the collector to a specific k8s cluster

# 1. VERSION (OPTIONAL)
# 2. NAMESPACE (OPTIONAL)
# 3. IMAGE(OPTIONAL) -- if missing build from source

DEFAULT_IMAGE_NAME="wavefronthq\/wavefront-kubernetes-collector"
DEFAULT_VERSION="1.2.6"

#TODO: use getopts instead?
WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3
IMAGE=$4
PROFILE=$5

echo "$WAVEFRONT_CLUSTER $API_TOKEN $VERSION $IMAGE $PROFILE"

if [[ -z ${WAVEFRONT_CLUSTER} || -z ${API_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    echo "wavefront cluster and token required"
    exit 1
fi

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi
NAMESPACE_VERSION=$(echo "$VERSION" | tr . -)

if [[ -z ${IMAGE} ]] ; then
    IMAGE=${DEFAULT_IMAGE_NAME}
fi

BASE_DIR="base"
OVERLAYS_DIR="overlays"

if [[ -z ${PROFILE} ]] ; then
    PROFILE=${OVERLAYS_DIR}/test
else
    PROFILE=${OVERLAYS_DIR}/${PROFILE}
fi

#TODO: temp directory for intermediate files
#TODO: need to replace the kustomize template to source from temp directory

#TODO: Migrate these sed to use kustomize edit
sed "s/YOUR_CLUSTER/${WAVEFRONT_CLUSTER}/g; s/YOUR_API_TOKEN/${API_TOKEN}/g" ${BASE_DIR}/proxy.template.yaml > ${BASE_DIR}/proxy.yaml
sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" ${BASE_DIR}/kustomization.template.yaml | sed "s/YOUR_IMAGE_TAG/${VERSION}/g" | sed "s/YOUR_IMAGE_NAME/${IMAGE}/g" > ${BASE_DIR}/kustomization.yaml

sed "s/YOUR_CLUSTER_NAME/cluster-${VERSION}/g" ${OVERLAYS_DIR}/test/collector.yaml.template | sed "s/NAMESPACE/${NAMESPACE_VERSION}-wavefront-collector/g" > ${OVERLAYS_DIR}/test/collector.yaml

kustomize build ${PROFILE} | kubectl apply -f -
