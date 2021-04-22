#!/usr/bin/env bash

NS=wavefront-collector
ROOT_DIR=$(git rev-parse --show-toplevel)
TEMP_DIR=$(mktemp -d)
CURRENT_VERSION=1.3.4
VERSION=1.3.5-rc2
WF_CLUSTER=nimba
WF_TOKEN=0ce14176-ce9e-4bc8-ade9-8d63567a5e52
K8S_CLUSTER=$VERSION-release-test

echo "Temp dir: $TEMP_DIR"

make nuke-kind
make deploy-targets

cp ${ROOT_DIR}/deploy/kubernetes/*  $TEMP_DIR/.

cp ${ROOT_DIR}/hack/deploy/memcached-config.yaml $TEMP_DIR/.
cp ${ROOT_DIR}/hack/deploy/mysql-config.yaml $TEMP_DIR/.
cp ${ROOT_DIR}/hack/deploy/prom-example.yaml $TEMP_DIR/.
cp ${ROOT_DIR}/hack/kustomize/base/proxy.template.yaml $TEMP_DIR/proxy.yaml

pushd $TEMP_DIR
  sed -i '' "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WF_TOKEN}/g" ${TEMP_DIR}/proxy.yaml
  sed -i '' "s/k8s-cluster/${K8S_CLUSTER}/g" ${TEMP_DIR}/4-collector-config.yaml
  sed -i '' "s/${CURRENT_VERSION}/${VERSION}/g" ${TEMP_DIR}/5-collector-daemonset.yaml

  kubectl config set-context --current --namespace=${NS}
  kubectl apply -f $TEMP_DIR/.
  kubectl config set-context --current --namespace=default
popd
