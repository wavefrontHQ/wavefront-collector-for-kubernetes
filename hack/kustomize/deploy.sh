#! /bin/bash

# This script automates the deployment of the collector to a specific k8s cluster

# 1. VERSION (OPTIONAL)
# 2. NAMESPACE (OPTIONAL)
# 3. IMAGE(OPTIONAL) -- if missing build from source

DEFAULT_VERSION="1.2.6"

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3

if [[ -z ${WAVEFRONT_CLUSTER} || -z ${API_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    echo "wavefront cluster and token required"
    exit 1
fi

#TODO: change the user provided cluster name that shows up in Wavefront

# TODO: input the base image
# input the diff image
# emit the output out to a log dump and then diff becomes easy

BASE_DIR="base"
OVERLAYS_DIR="overlays"

#TODO: temp directory for intermediate files
#TODO: need to replace the kustomize template to source from temp directory

sed "s/YOUR_CLUSTER/${WAVEFRONT_CLUSTER}/g; s/YOUR_API_TOKEN/${API_TOKEN}/g" ${BASE_DIR}/proxy.template.yaml > ${BASE_DIR}/proxy.yaml

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

sed "s/YOUR_IMAGE_TAG/${VERSION}/g" ${OVERLAYS_DIR}/versions/kustomization.template.yaml > ${OVERLAYS_DIR}/versions/kustomization.yaml
sed "s/YOUR_CLUSTER_NAME/cluster-${VERSION}/g" ${OVERLAYS_DIR}/test/config/collector.yaml.template > ${OVERLAYS_DIR}/test/config/collector.yaml

cat ${BASE_DIR}/proxy.yaml
cat ${OVERLAYS_DIR}/versions/kustomization.yaml

#TODO: figure out how to deploy to a cluster and which cluster

kubectl apply -k ${OVERLAYS_DIR}/test/config/
