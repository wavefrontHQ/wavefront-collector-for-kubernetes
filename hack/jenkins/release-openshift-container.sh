#!/bin/bash -e

PREFIX=$1
REDHAT_CREDS_USR=$2
REDHAT_CREDS_PSW=$3
REDHAT_API_KEY=$4
REDHAT_PROJECT_ID=$5
GIT_BRANCH=$6
RC_NUMBER=$7

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

pwd
VERSION=$(cat release/VERSION)-rc${RC_NUMBER}
echo VERSION is ${VERSION}
#podman login ${PREFIX} -u ${REDHAT_CREDS_USR} -p ${REDHAT_CREDS_PSW}
#podman build -f deploy/docker/Dockerfile-rhel --build-arg=COLLECTOR_VERSION=${VERSION} -t ${PREFIX}/wavefront:${VERSION} .
#podman push ${PREFIX}/wavefront:${VERSION}
#export PFLT_DOCKERCONFIG=${XDG_RUNTIME_DIR}/containers/auth.json
#preflight check container ${PREFIX}/wavefront:${VERSION} --pyxis-api-token=${REDHAT_API_KEY}
#preflight check container ${PREFIX}/wavefront:${VERSION} --pyxis-api-token=${REDHAT_API_KEY} --submit --certification-project-id=${REDHAT_PROJECT_ID}