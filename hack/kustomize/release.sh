#! /bin/bash

# This script automates the deployment of the collector to a specific k8s cluster

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2

if [[ -z ${WAVEFRONT_CLUSTER} || -z ${API_TOKEN} ]] ; then
    echo "wavefront cluster and token required"
    exit 1
fi

ODIR="overlays/common"

sed "s/YOUR_CLUSTER/${WAVEFRONT_CLUSTER}/g; s/YOUR_API_TOKEN/${API_TOKEN}/g" ${ODIR}/proxy.template.yaml > ${ODIR}/proxy.yaml

cat ${ODIR}/proxy.yaml
