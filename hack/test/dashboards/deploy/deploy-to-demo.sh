#!/usr/bin/env bash
set -e

cd /home/k8po/wavefront-collector-for-kubernetes/hack/test/dashboards/deploy/
sed -i "s/YOUR_CLUSTER/${WF_CLUSTER}/g; s/YOUR_API_TOKEN/${WAVEFRONT_TOKEN}/g" "6-wavefront-proxy.yaml"

kubectl delete -f . || true
kubectl apply -f .