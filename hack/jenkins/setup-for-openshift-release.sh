#!/bin/bash -e

#
# preflight
#
if ! [ -x "$(command -v preflight)" ]; then
    curl -LO https://github.com/redhat-openshift-ecosystem/openshift-preflight/releases/download/1.1.1/preflight-linux-amd64
    ls
fi
