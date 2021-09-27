#!/bin/bash -xe

MESSAGE="wavefront-collector-for-kubernetes:v${BUILD_VERSION}"


# tobs-k8s-assist channel ID: C01BXLYMB3K
# changed this as per ticket ESO-3126
echo curl -X POST --data-urlencode "payload={\"channel\": \"${CHANNEL_ID}\", \"username\": \"jenkins\", \"text\": \"Success!! \`${MESSAGE}\` released by ${BUILD_USER}(${BUILD_USER_ID})!\"}" ${SLACK_WEBHOOK_URL}

