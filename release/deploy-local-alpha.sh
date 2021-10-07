#!/usr/bin/env bash
set -e

cd "$(dirname "$0")" # cd to deploy-local-alpha.sh is in

source "../hack/deploy/k8s-utils.sh"

if [[ -z ${CURRENT_VERSION} ]] ; then
    print_msg_and_exit "Need to specify alpha version image tag by setting CURRENT_VERSION"
fi

CURRENT_COLLECTOR_REPO=projects.registry.vmware.com/tanzu_observability_keights_saas/collector-snapshot \
./deploy-local.sh