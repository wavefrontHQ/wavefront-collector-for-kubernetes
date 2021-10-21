#!/bin/bash -xe

BUILD_VERSION=$(cat ./release/VERSION)

## tobs-k8s-assist channel
#curl -X POST --data-urlencode "payload={\"channel\": \"${CHANNEL_ID}\", \"username\": \"jenkins\", \"text\": \"Success!! \`${MESSAGE}\` released by ${BUILD_USER}(${BUILD_USER_ID})!\"}" ${SLACK_WEBHOOK_URL}
#curl --fail -X POST -H "Content-Type:application/json" \
#-H "Authorization: token $GITHUB_CREDS_PSW" \
#-d "{
#      \"tag_name\": \"v$VERSION\",
#      \"target_commitish\": \"$GIT_BRANCH\",
#      \"name\": \"Release v$VERSION\",
#      \"body\": \"Description for v$VERSION\",
#      \"draft\": true,
#      \"prerelease\": false}" \
#"https://api.github.com/repos/$GIT_HUB_REPO/releases"

curl -u $HARBOR_CREDS_USR:$HARBOR_CREDS_PSW \
-X POST "https://projects.registry.vmware.com/api/v2.0/projects/tanzu_observability/repositories/kubernetes-collector/artifacts/1.3.5-rc2/tags" \
-H "accept: application/json" \
-H "Content-Type: application/json" \
-d "{
      \"name\": \"1.3.5-rc2-test\",
      \"signed\": true,
      \"immutable\": true}"
