NS=$(kubectl get namespaces | awk '/collector-targets/ {print $1}')
if [ -z ${NS} ]; then exit 0; fi

echo "Uninstalling targets..."
kubectl delete -f prom-example.yaml &>/dev/null || true
kubectl delete -f exclude-prom-example.yaml &>/dev/null || true
kubectl delete -f cpu-throttled-prom-example.yaml &>/dev/null || true
kubectl delete -f pod-cannot-be-schedule.yaml &>/dev/null || true

helm uninstall memcached-release --namespace collector-targets &>/dev/null || true

helm uninstall mysql-release --namespace collector-targets &>/dev/null || true

kubectl delete namespace collector-targets &>/dev/null || true

echo "Targets uninstalled"