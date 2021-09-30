#!/usr/bin/env bash
set -ex

cd "$(dirname "$0")" # cd to directory that bump-version.sh is in

pushd ../
  make semver-cli
  export PATH=$PATH:$GOPATH/bin
popd

BUMP_COMPONENT=$1

if [[ -z "${BUMP_COMPONENT}" ]] ; then
    echo "usage: ./release/bump-version.sh <semver component to bump>"
    exit 1
fi

DEPLOY_DIR=../deploy

OLD_VERSION=$(cat ./VERSION)
NEXT_VERSION=$(semver-cli inc "$BUMP_COMPONENT" "$OLD_VERSION")

GIT_BRANCH="bump-${NEXT_VERSION}"
git checkout -b "$GIT_BRANCH"
echo "${GIT_BRANCH}" > ./GIT_BUMP_BRANCH_NAME

## Bump to next version
sed -i "" "s/${OLD_VERSION}/${NEXT_VERSION}/g" "$DEPLOY_DIR/kubernetes/5-collector-daemonset.yaml"
sed -i "" "s/${OLD_VERSION}/${NEXT_VERSION}/g" "$DEPLOY_DIR/openshift/collector/3-collector-deployment.yaml"
echo "$NEXT_VERSION" > ./VERSION

git commit -am "bump version to ${NEXT_VERSION}"
git push --set-upstream origin "$GIT_BRANCH"

gh pr create --base master --fill --head "$GIT_BRANCH" --web