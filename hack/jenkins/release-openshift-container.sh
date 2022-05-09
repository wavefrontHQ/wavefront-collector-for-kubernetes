#!/bin/bash -e

#
# preflight
#
if ! [ -x "$(command -v preflight)" ]; then
    curl -LO https://github.com/redhat-openshift-ecosystem/openshift-preflight/releases/download/1.1.1/preflight-linux-amd64
    chmod +x ./preflight-linux-amd64
    sudo mv ./preflight-linux-amd64 /usr/local/bin/preflight
fi

cd workspace/wavefront-collector-for-kubernetes/
git clean -dfx
git checkout ${GIT_BRANCH}
git pull

podman login ${PREFIX} -u ${REDHAT_CREDS_USR} -p ${REDHAT_CREDS_PSW}
podman build -f deploy/docker/Dockerfile-rhel --build-arg=COLLECTOR_VERSION=1.11.0 -t ${PREFIX}/wavefront:1.11.0 .
podman push ${PREFIX}/wavefront:1.11.0
export PFLT_DOCKERCONFIG=${XDG_RUNTIME_DIR}/containers/auth.json
preflight check container ${PREFIX}/wavefront:1.11.0 --pyxis-api-token=${REDHAT_API_KEY}
preflight check container ${PREFIX}/wavefront:1.11.0 --pyxis-api-token=${REDHAT_API_KEY} --submit --certification-project-id=${REDHAT_PROJECT_ID}