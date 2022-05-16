#!/bin/bash -e
source ./deploy/k8s-utils.sh

REPO_ROOT=$(git rev-parse --show-toplevel)

cd "$(dirname "$0")" # hack/test

DEFAULT_VERSION="1.10.0"
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
    echo -e "\t-y collector yaml"
    exit 1
}

WF_CLUSTER=
WAVEFRONT_TOKEN=
VERSION=
K8S_ENV=gke
COLLECTOR_YAML=

while getopts "c:t:v:d:k:y:" opt; do
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
    y)
      COLLECTOR_YAML="$OPTARG"
      ;;
    \?)
      print_usage_and_exit "Invalid option: -$OPTARG"
      ;;
  esac
done

if [[ -z ${WF_CLUSTER} || -z ${WAVEFRONT_TOKEN} ]] ; then
    #TODO: source these from environment variables if not provided
    print_usage_and_exit "wavefront instance and token required"
fi


if [[ -z ${VERSION} ]] ; then
    VERSION=${DEFAULT_VERSION}
fi

if [[ -z ${COLLECTOR_YAML} ]] ; then
    COLLECTOR_YAML="${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml"
fi

cp "${REPO_ROOT}/deploy/kubernetes/0-collector-namespace.yaml" base/deploy/0-collector-namespace.yaml
cp "${REPO_ROOT}/deploy/kubernetes/1-collector-cluster-role.yaml" base/deploy/1-collector-cluster-role.yaml
cp "${REPO_ROOT}/deploy/kubernetes/2-collector-rbac.yaml" base/deploy/2-collector-rbac.yaml
cp "${REPO_ROOT}/deploy/kubernetes/3-collector-service-account.yaml" base/deploy/3-collector-service-account.yaml
cp "${REPO_ROOT}/deploy/kubernetes/5-collector-daemonset.yaml" base/deploy/collector-deployments/5-collector-daemonset.yaml
cp "${REPO_ROOT}/deploy/kubernetes/alternate-collector-deployments/5-collector-single-deployment.yaml" base/deploy/collector-deployments/5-collector-single-deployment.yaml
cp "${REPO_ROOT}/deploy/kubernetes/alternate-collector-deployments/5-collector-combined.yaml" base/deploy/collector-deployments/5-collector-collector-combined.yaml

cp "${COLLECTOR_YAML}" base/deploy/5-wavefront-collector.yaml


NS=wavefront-collector
WF_CLUSTER_NAME=$(whoami)-${K8S_ENV}-$(date +"%y%m%d")

if $USE_TEST_PROXY ; then
  cp base/test-proxy.yaml base/deploy/6-wavefront-proxy.yaml
else
  sed "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" base/proxy.template.yaml > base/deploy/6-wavefront-proxy.yaml
fi

# TODO: only sed into kustomization template and have it fill in variables in files
 sed "s/YOUR_IMAGE_TAG/${VERSION}/g" base/kustomization.template.yaml  > base/kustomization.yaml

sed "s/YOUR_CLUSTER_NAME/${WF_CLUSTER_NAME}/g"  base/collector-config.template.yaml  |
  sed "s/NAMESPACE/${NS}/g" |
  sed "s/FLUSH_INTERVAL/${FLUSH_INTERVAL}/g" |
  sed  "s/COLLECTION_INTERVAL/${COLLECTION_INTERVAL}/g" > base/deploy/4-collector-config.yaml


if  [ ${USE_TEST_PROXY} == false ] && [ "${VERSION}" != "fake" ]; then
  echo "IMAGE TAG: ${VERSION}"
  green "WF Cluster Name: ${WF_CLUSTER_NAME}"
fi