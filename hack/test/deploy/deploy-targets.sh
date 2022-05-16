#! /bin/bash -e

SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/k8s-utils.sh"

cd "$SCRIPT_DIR"

echo "Deploying targets..."

kubectl create namespace collector-targets &> /dev/null || true

while ! kubectl get --namespace collector-targets serviceaccount/default  &> /dev/null; do
    echo "Waiting for namespace service account to be created'"
    sleep 1
done

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

wait_for_cluster_ready
echo "Finished deploying targets"