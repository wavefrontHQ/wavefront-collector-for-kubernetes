#!/bin/bash -e

cd "$(dirname "$0")" # cd to directory that bump-version.sh is in

source ./VERSION

KUSTOMIZE_DIR=../hack/kustomize
DEPLOY_DIR=../deploy
TMP_FILE=/tmp/temporary

GIT_BRANCH="bump-${CURRENT_VERSION}"
git checkout -b $GIT_BRANCH

## Bump to current version
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/kubernetes/5-collector-daemonset.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${DEPLOY_DIR}/openshift/collector/3-collector-deployment.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/base/5-collector-daemonset.yaml
sed -i "" "s/${RELEASED_VERSION}/${CURRENT_VERSION}/g" ${KUSTOMIZE_DIR}/deploy.sh

## Bump to future version
sed -i "" "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ../Makefile
sed -i "" "s/${CURRENT_VERSION}/${NEXT_VERSION}/g" ${KUSTOMIZE_DIR}/test.sh

git commit -am "bump version to ${CURRENT_VERSION}"
git push --set-upstream origin $GIT_BRANCH

gh pr create --base master --fill --head $GIT_BRANCH --web