#!/usr/bin/env bash
set -e

cd "$(dirname "$-1")"

VERSION=$(cat ./release/VERSION)
GIT_HUB_REPO=wavefrontHQ/wavefront-collector-for-kubernetes
GIT_BRANCH=master #update when we get choosing a specific branch working


curl -X POST -H "Content-Type:application/json" \
-H "Authorization: token $GITHUB_CREDS_PSW" \
-d "{
      \"tag_name\": \"v$VERSION\",
      \"target_commitish\": \"$GIT_BRANCH\",
      \"name\": \"Release v$VERSION\",
      \"body\": \"Description for v$VERSION\",
      \"draft\": true,
      \"prerelease\": false}" \
"https://api.github.com/repos/$GIT_HUB_REPO/releases"