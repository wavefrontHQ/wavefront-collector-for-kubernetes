#!/usr/bin/env bash
source hack/make/_script-tools.sh

if [[ -z ${K8S_ENV} ]]; then
  print_msg_and_exit 'K8S_ENV required but was empty'
  #K8S_ENV=$DEFAULT_K8S_ENV
fi

if [[ -z ${KUSTOMIZE_DIR} ]]; then
  print_msg_and_exit 'KUSTOMIZE_DIR required but was empty'
  #KUSTOMIZE_DIR=$DEFAULT_KUSTOMIZE_DIR
fi

if [[ -z ${WAVEFRONT_TOKEN} ]]; then
  print_msg_and_exit 'WAVEFRONT_TOKEN required but was empty'
  #WAVEFRONT_TOKEN=$DEFAULT_WAVEFRONT_TOKEN
fi

if [[ -z ${VERSION} ]]; then
  print_msg_and_exit 'VERSION required but was empty'
  #VERSION=$DEFAULT_VERSION
fi

if [[ -z ${GCP_PROJECT} ]]; then
  print_msg_and_exit 'GCP_PROJECT required but was empty'
  #GCP_PROJECT=$DEFAULT_GCP_PROJECT
fi
WAVEFRONT_DEV_AWS_ACC_ID=095415062695
AWS_PROFILE=wavefront-dev
AWS_REGION=us-west-2
ECR_ENDPOINT=${WAVEFRONT_DEV_AWS_ACC_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
# commands ...
pushd_check ${KUSTOMIZE_DIR}
if [ ${K8S_ENV} == "GKE" ]; then
  ./test.sh nimba ${WAVEFRONT_TOKEN} ${VERSION} "us.gcr.io\/${GCP_PROJECT}"
elif [ ${K8S_ENV} == "EKS" ]; then
  echo "${ECR_ENDPOINT}\/tobs/k8s/saas"
  echo "us.gcr.io\/${GCP_PROJECT}"
  ./test.sh nimba ${WAVEFRONT_TOKEN} ${VERSION} "${ECR_ENDPOINT}\/tobs\/k8s\/saas"
else
  ./test.sh nimba ${WAVEFRONT_TOKEN} ${VERSION}
fi
popd_check ${KUSTOMIZE_DIR}
