#! /bin/bash -e

SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/k8s-utils.sh"

cd "$SCRIPT_DIR"

echo "Deploying targets..."

kubectl create namespace collector-targets &> /dev/null || true

kubectl apply -f prom-example.yaml &>/dev/null || true
kubectl apply -f exclude-prom-example.yaml &>/dev/null || true
kubectl apply -f cpu-throttled-prom-example.yaml &>/dev/null || true
kubectl apply -f pending-pod-cannot-be-scheduled.yaml &>/dev/null || true
kubectl apply -f pending-pod-image-cannot-be-loaded.yaml &>/dev/null || true

kubectl delete -f jobs.yaml &>/dev/null || true
kubectl apply -f jobs.yaml &>/dev/null || true

helm repo add bitnami https://charts.bitnami.com/bitnami &>/dev/null || true
helm upgrade --install memcached-release bitnami/memcached \
--set resources.requests.memory="100Mi",resources.requests.cpu="100m" \
--set persistence.size=200Mi \
--namespace collector-targets &>/dev/null || true

helm upgrade --install mysql-release bitnami/mysql \
--set auth.rootPassword=password123 \
--set primary.persistence.size=200Mi \
--namespace collector-targets &>/dev/null || true

wait_for_cluster_ready
echo "Finished deploying targets"