#! /bin/bash -e


NS=$(kubectl get namespaces | awk '/wavefront-collector/ {print $1}')


if [ -z ${NS} ]; then exit 0; fi

echo "deleting wavefront collector deployment"
kustomize build overlays/test | kubectl delete -f - || true