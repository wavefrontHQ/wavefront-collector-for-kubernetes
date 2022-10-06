#!/bin/bash -e

REPO_ROOT=$(git rev-parse --show-toplevel)
source ${REPO_ROOT}/hack/test/deploy/k8s-utils.sh

function main() {
  rm $REPO_ROOT/hack/integrations/working/*.partial*
  rm $REPO_ROOT/hack/integrations/working/*-partial*
}

main $@
