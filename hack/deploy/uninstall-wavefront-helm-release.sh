RELEASE=$(helm list -n wavefront | awk '/wavefront/ {print $1}')
if [ -z ${RELEASE} ]; then exit 0; fi

echo "uninstalling wavefront helm release"
helm uninstall wavefront --namespace wavefront || true
kubectl delete namespace wavefront || true