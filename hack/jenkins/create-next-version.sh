#!/usr/bin/env bash
set -ex

cd "$(dirname "$0")" # cd to directory that bump-version.sh is in

pushd ../../
  make semver-cli
popd

BUMP_COMPONENT=$1

if [[ -z "${BUMP_COMPONENT}" ]] ; then
    echo "usage: ./release/bump-version.sh <semver component to bump>"
    exit 1
fi


OLD_VERSION=$(cat ../../release/VERSION)
echo "${OLD_VERSION}" > ./OLD_VERSION
NEXT_VERSION=$(semver-cli inc "$BUMP_COMPONENT" "$OLD_VERSION")
echo "${NEXT_VERSION}" > ./NEXT_VERSION

GIT_BRANCH="bump-${NEXT_VERSION}"
git checkout "$GIT_BRANCH" 2>/dev/null || git checkout -b "$GIT_BRANCH"
echo "${GIT_BRANCH}" > ./GIT_BUMP_BRANCH_NAME
