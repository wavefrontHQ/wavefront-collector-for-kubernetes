#!/usr/bin/env bash
set -ex

REPO_ROOT=$(git rev-parse --show-toplevel)
cd $REPO_ROOT

make semver-cli

BUMP_COMPONENT=$1

if [[ -z "${BUMP_COMPONENT}" ]] ; then
    echo "usage: ./hack/jenkins/create-bump-version-branch.sh <semver component to bump>"
    exit 1
fi

OLD_VERSION=$(cat ./release/VERSION)
NEW_VERSION=$(semver-cli inc "$BUMP_COMPONENT" "$OLD_VERSION")
echo "$NEW_VERSION" >./release/VERSION

GIT_BUMP_BRANCH_NAME="bump-${NEW_VERSION}"
git checkout -b "$GIT_BUMP_BRANCH_NAME"

echo "Bumping ${OLD_VERSION} to ${NEW_VERSION} in ./deploy/kubernetes/5-collector-daemonset.yaml"
sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" "./deploy/kubernetes/5-collector-daemonset.yaml"
git commit -am "bump version to ${NEW_VERSION}"

git push --set-upstream origin "${GIT_BUMP_BRANCH_NAME}"
