#! /bin/bash

# This script automates the functional testing of the collector

DEFAULT_VERSION="1.2.7-beta1"
DEFAULT_IMAGE_NAME="vikramraman\/wavefront-collector"

WAVEFRONT_CLUSTER=$1
API_TOKEN=$2

LOG_FILE="/tmp/logs.txt"
SORTED_FILE="/tmp/sorted.txt"

echo "deploying collector ${DEFAULT_IMAGE_NAME} ${DEFAULT_VERSION}"
./deploy.sh ${WAVEFRONT_CLUSTER} ${API_TOKEN} ${DEFAULT_VERSION} ${DEFAULT_IMAGE_NAME} flushonce

echo "waiting for logs to be emitted"
sleep 30

NAMESPACE_VERSION=$(echo "${DEFAULT_VERSION}" | tr . -)
NS=${NAMESPACE_VERSION}-wavefront-collector

PODS=`kubectl -n ${NS} get pod -l k8s-app=wavefront-collector | awk '{print $1}' | tail +2`

rm -f ${LOG_FILE}

# TODO: relies on the prefix from the sample app to isolate prom metrics
PROM_PREFIX="prom-example."

for pod in ${PODS} ; do
    echo "writing metrics from ${pod} into ${LOG_FILE}"
    kubectl logs ${pod} -n ${NS} | grep 'Metric: prom-example' | jq -r '.msg' >> ${LOG_FILE}
done

sort ${LOG_FILE} > ${SORTED_FILE}

echo "diff between prom files"
echo ${SORTED_FILE}
diff -q ${SORTED_FILE} files/prometheus-baseline.txt
if [[ $? -eq 0 ]] ; then
   echo "diff succeeded"
fi
