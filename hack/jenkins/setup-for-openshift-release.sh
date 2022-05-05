#!/bin/bash -e

#
# preflight
#
if ! [ -x "$(command -v preflight)" ]; then
    curl -LO https://github.com/redhat-openshift-ecosystem/openshift-preflight/releases/download/1.1.1/preflight-linux-amd64
    chmod +x ./preflight-linux-amd64
    sudo mv ./preflight-linux-amd64 /usr/local/bin/preflight
fi
