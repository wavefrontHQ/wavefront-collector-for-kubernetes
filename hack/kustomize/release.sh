#! /bin/bash

# This script automates the deployment of the collector to a specific k8s cluster

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3

if [[ -z ${WAVEFRONT_CLUSTER} || -z ${API_TOKEN} ]] ; then
    echo "wavefront cluster and token required"
    exit 1
fi

# TODO: input the base image
# input the diff image
# emit the output out to a log dump and then diff becomes easy

ODIR="overlays/common"

sed "s/YOUR_CLUSTER/${WAVEFRONT_CLUSTER}/g; s/YOUR_API_TOKEN/${API_TOKEN}/g" ${ODIR}/proxy.template.yaml > ${ODIR}/proxy.yaml

if [[ -z ${VERSION} ]] ; then
    VERSION="1.2.7" 
fi

sed "s/YOUR_IMAGE_TAG/${VERSION}/g" ${ODIR}/../versions/kustomization.template.yaml > ${ODIR}/../versions/kustomization.yaml

cat ${ODIR}/proxy.yaml
cat ${ODIR}/../versions/kustomization.yaml
