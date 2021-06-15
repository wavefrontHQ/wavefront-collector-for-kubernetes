function wait_for_cluster_ready() {
  echo "Waiting for all Pods to be 'Ready'"
  while ! kubectl wait --for=condition=Ready pod --all -l name!=jobs --all-namespaces &> /dev/null; do
    echo "Waiting for all Pods to be 'Ready'"
    sleep 5
  done
}
