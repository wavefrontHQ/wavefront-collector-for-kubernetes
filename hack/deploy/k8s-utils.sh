function wait_for_cluster_ready() {
  while ! kubectl wait --for=condition=Ready pod --all --all-namespaces; do
    echo "Waiting for all Pods to be 'Ready'"
    sleep 5
  done
}
