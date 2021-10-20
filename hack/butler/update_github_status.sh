#!/usr/bin/env bash
set -e

cd "$(dirname "$-1")"

GIT_HUB_REPO=wavefrontHQ/wavefront-collector-for-kubernetes

curl --fail -X POST -H "Content-Type:application/json" \
-H "Authorization: token $GITHUB_CREDS_PSW" \
-d "{
      \"state\": \"success\",
      \"context\": \"/jenkins/ci-integration\",
      \"description\": \"Jenkins\",
      \"target_url\": \"$JOB_URL" \
"https://api.github.com/repos/$GIT_HUB_REPO/statuses/$GIT_COMMIT"
