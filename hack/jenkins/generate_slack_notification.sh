#!/bin/bash -xe

BUILD_VERSION=$(cat ./release/VERSION)
MESSAGE="wavefront-collector-for-kubernetes:v${BUILD_VERSION}"

# tobs-k8s-assist channel
curl -X POST --data-urlencode "payload={\"channel\": \"${CHANNEL_ID}\", \"username\": \"jenkins\", \"text\": \"Success!! \`${MESSAGE}\` released by ${BUILD_USER}(${BUILD_USER_ID})!\"}" ${SLACK_WEBHOOK_URL}

