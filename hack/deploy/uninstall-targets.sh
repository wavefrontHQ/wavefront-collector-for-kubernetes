NS=$(kubectl get namespaces | awk '/collector-targets/ {print $1}')
if [ -z ${NS} ]; then exit 0; fi

echo "deleting prometheus endpoint"
kubectl delete -f prom-example.yaml || true

echo "uninstalling memcached"
helm uninstall memcached-release --namespace collector-targets || true

echo "uninstalling mysql"
helm uninstall mysql-release --namespace collector-targets || true

echo "deleting  namespace collector-targets"
kubectl delete namespace collector-targets || true

echo "targets uninstalled"