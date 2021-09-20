#!/usr/bin/env bash
set -e

GIT_BRANCH=move-to-butler
VERSION=test
GIT_REPO_NAME=wavefronthq/wavefront-collector-for-kubernetes

curl -X POST -H "Content-Type:application/json" \
-H "Authorization: token ${GITHUB_CREDS_PSW}" \
-d "{
      \"tag_name\": \"v$VERSION\",
      \"target_commitish\": \"$GIT_BRANCH\",
      \"name\": \"Release v$VERSION\",
      \"body\": \"Description for v$VERSION\",
      \"draft\": true,
      \"prerelease\": false}" \
"https://api.github.com/repos/$GIT_REPO_NAME/releases"