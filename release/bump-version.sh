#!/bin/bash -e

RELEASED_VERSION="1.3.1"
CURRENT_VERSION="1.3.2"
NEXT_VERSION="1.3.3"

KUSTOMIZE_DIR=../hack/kustomize
DEPLOY_DIR=../deploy
TMP_FILE=/tmp/temporary

## Bump to current version
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/kubernetes/5-collector-daemonset.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/openshift/collector/3-collector-deployment.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/base/5-collector-daemonset.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/deploy.sh

## Bump to future version
sed -i "" "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ../Makefile
sed -i "" "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ${KUSTOMIZE_DIR}/test.sh