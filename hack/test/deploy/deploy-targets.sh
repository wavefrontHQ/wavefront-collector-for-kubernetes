#! /bin/bash -e

SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/k8s-utils.sh"

cd "$SCRIPT_DIR"

echo "Deploying targets..."

kubectl delete namespace collector-targets &> /dev/null || true

wait_for_namespace_created collector-targets

wait_for_namespaced_resource_created collector-targets serviceaccount/default

kubectl apply -f prom-example.yaml >/dev/null
kubectl apply -f exclude-prom-example.yaml >/dev/null
kubectl apply -f cpu-throttled-prom-example.yaml >/dev/null
kubectl apply -f pending-pod-cannot-be-scheduled.yaml >/dev/null
kubectl apply -f pending-pod-image-cannot-be-loaded.yaml >/dev/null

kubectl delete -f jobs.yaml &>/dev/null || true
kubectl apply -f jobs.yaml >/dev/null

helm repo add bitnami https://charts.bitnami.com/bitnami &>/dev/null || true
helm upgrade --install memcached-release bitnami/memcached \
--set resources.requests.memory="100Mi",resources.requests.cpu="100m" \
--set persistence.size=200Mi \
--namespace collector-targets >/dev/null

helm upgrade --install mysql-release bitnami/mysql \
--set auth.rootPassword=password123 \
--set primary.persistence.size=200Mi \
--namespace collector-targets >/dev/null

echo "Finished deploying targets"