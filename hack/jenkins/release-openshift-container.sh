#!/bin/bash -e

PREFIX=$1
REDHAT_CREDS_USR=$2
REDHAT_CREDS_PSW=$3
REDHAT_API_KEY=$4
REDHAT_PROJECT_ID=$5
GIT_BUMP_BRANCH_NAME=$6
RC_NUMBER=$7

if ! [ -x "$(command -v preflight)" ]; then
    curl -LO https://github.com/redhat-openshift-ecosystem/openshift-preflight/releases/download/1.1.1/preflight-linux-amd64
    chmod +x ./preflight-linux-amd64
    sudo mv ./preflight-linux-amd64 /usr/local/bin/preflight
fi

cd /root/workspace/wavefront-collector-for-kubernetes/
git clean -dfx
git checkout -- .
git checkout main
git pull
git checkout ${GIT_BUMP_BRANCH_NAME}
VERSION=$(cat release/VERSION)
TAG_VERSION=${VERSION}-rc${RC_NUMBER}

podman login ${PREFIX} -u ${REDHAT_CREDS_USR} -p ${REDHAT_CREDS_PSW}
podman build -f deploy/docker/Dockerfile-rhel --build-arg=COLLECTOR_VERSION=${VERSION} -t ${PREFIX}/wavefront:${TAG_VERSION} .
podman push ${PREFIX}/wavefront:${TAG_VERSION}
export PFLT_DOCKERCONFIG=${XDG_RUNTIME_DIR}/containers/auth.json
preflight check container ${PREFIX}/wavefront:${TAG_VERSION} --pyxis-api-token=${REDHAT_API_KEY}
preflight check container ${PREFIX}/wavefront:${TAG_VERSION} --pyxis-api-token=${REDHAT_API_KEY} --submit --certification-project-id=${REDHAT_PROJECT_ID}

# At wrap up, delete local bump branch to keep Openshift VM clean
git checkout main
git branch -D ${GIT_BUMP_BRANCH_NAME}