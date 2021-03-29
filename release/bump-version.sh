#!/bin/bash -e

RELEASED_VERSION="1.3.1"
CURRENT_VERSION="1.3.2"
NEXT_VERSION="1.3.3"

KUSTOMIZE_DIR=../hack/kustomize
DEPLOY_DIR=../deploy
TMP_FILE=/tmp/temporary

## Bump to current version
sed "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/kubernetes/5-collector-daemonset.yaml > ${TMP_FILE} && mv ${TMP_FILE} ${DEPLOY_DIR}/kubernetes/5-collector-daemonset.yaml
sed "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/openshift/collector/3-collector-deployment.yaml > ${TMP_FILE} && mv ${TMP_FILE} ${DEPLOY_DIR}/openshift/collector/3-collector-deployment.yaml
sed "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/base/5-collector-daemonset.yaml > ${TMP_FILE} && mv ${TMP_FILE} ${KUSTOMIZE_DIR}/base/5-collector-daemonset.yaml
sed "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/deploy.sh > ${TMP_FILE} && mv ${TMP_FILE} ${KUSTOMIZE_DIR}/deploy.sh

## Bump to future version
sed "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ../Makefile > ${TMP_FILE} && mv ${TMP_FILE} ../Makefile
sed "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ${KUSTOMIZE_DIR}/test.sh > ${TMP_FILE} && mv ${TMP_FILE} ${KUSTOMIZE_DIR}/test.sh