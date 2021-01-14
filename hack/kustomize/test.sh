#! /bin/bash

# This script automates the deployment of the collector to a specific k8s cluster

# 1. VERSION (OPTIONAL)
# 2. NAMESPACE (OPTIONAL)
# 3. IMAGE(OPTIONAL) -- if missing build from source

DEFAULT_VERSION="1.2.7-beta1"
DEFAULT_IMAGE_NAME="vikramraman\/wavefront-collector"

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2
VERSION=$3

LOG_FILE="/tmp/logs.txt"

if [[ -z ${WAVEFRONT_CLUSTER} || -z ${API_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    echo "wavefront cluster and token required"
    exit 1
fi

if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

NAMESPACE_VERSION=$(echo "$VERSION" | tr . -)
NS=${NAMESPACE_VERSION}-wavefront-collector

# TODO: input the base image
# input the diff image
IMAGE_NAME=${DEFAULT_IMAGE_NAME}

BASE_DIR="base"
OVERLAYS_DIR="overlays"

#TODO: temp directory for intermediate files
#TODO: need to replace the kustomize template to source from temp directory

sed "s/YOUR_CLUSTER/${WAVEFRONT_CLUSTER}/g; s/YOUR_API_TOKEN/${API_TOKEN}/g" ${BASE_DIR}/proxy.template.yaml > ${BASE_DIR}/proxy.yaml

sed "s/NAMESPACE/${NS}/g" ${BASE_DIR}/kustomization.template.yaml | sed "s/YOUR_IMAGE_TAG/${VERSION}/g" | sed "s/YOUR_IMAGE_NAME/${IMAGE_NAME}/g" > ${BASE_DIR}/kustomization.yaml

#sed "s/YOUR_CLUSTER_NAME/cluster-${VERSION}/g" ${OVERLAYS_DIR}/test/config/collector.yaml.template > ${OVERLAYS_DIR}/test/config/collector.yaml

cat ${BASE_DIR}/proxy.yaml
cat ${OVERLAYS_DIR}/versions/kustomization.yaml

kustomize build ${OVERLAYS_DIR}/flushonce | kubectl apply -f -

# emit the output out to a log dump and then diff becomes easy
#TODO: diff the metrics

echo "waiting for logs to be emitted"
sleep 30

PODS=`kubectl -n ${NS} get pod -l k8s-app=wavefront-collector | awk '{print $1}' | tail +2`

rm -f ${LOG_FILE}

for pod in ${PODS} ; do
    echo "writing metrics from ${pod} into ${LOG_FILE}"
    kubectl logs ${pod} -n ${NS} | grep 'Metric:' | jq -r '.msg' >> ${LOG_FILE}
done
