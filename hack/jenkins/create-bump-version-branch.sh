#!/usr/bin/env bash
set -ex

cd "$(dirname "$0")" # cd to directory that create-bump-version-branch.sh is in

pushd ../../
  make semver-cli
popd

BUMP_COMPONENT=$1

if [[ -z "${BUMP_COMPONENT}" ]] ; then
    echo "usage: ./release/bump-version.sh <semver component to bump>"
    exit 1
fi


OLD_VERSION=$(cat ../../release/VERSION)
NEW_VERSION=$(semver-cli inc "$BUMP_COMPONENT" "$OLD_VERSION")
echo "$NEW_VERSION" >../../release/VERSION

GIT_BUMP_BRANCH_NAME="bump-${NEW_VERSION}"
git checkout -b "$GIT_BUMP_BRANCH_NAME"

echo "Bumping ${OLD_VERSION} to ${NEW_VERSION} in ../../deploy/kubernetes/5-collector-daemonset.yaml"
sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" "../../deploy/kubernetes/5-collector-daemonset.yaml"
git commit -am "bump version to ${NEW_VERSION}"

git push --set-upstream origin "${GIT_BUMP_BRANCH_NAME}"
