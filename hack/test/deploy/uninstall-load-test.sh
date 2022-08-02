#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh


function main() {
	cd "$(dirname "$0")"

	./uninstall-deploy-local.sh || true
	kubectl delete namespace scale-test || true
}

main $@
