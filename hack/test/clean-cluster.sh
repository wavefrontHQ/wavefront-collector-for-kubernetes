#!/usr/bin/env bash
set -e

echo "Cleaning up cluster"

CLUSTER_ROLES=$(kubectl get clusterroles | awk '/wavefront-collector|wavefront|wavefront-wavefront-collector|wavefront-wavefront-legacy-install-detection|wavefront-wavefront-logging/ {print $1}')
if [[ ! -z "$CLUSTER_ROLES" ]] ; then
    echo "Found ClusterRoles: ${CLUSTER_ROLES}"
    kubectl delete --wait=false clusterroles ${CLUSTER_ROLES} || true
	  kubectl delete --wait=false clusterrolebindings ${CLUSTER_ROLES} || true
fi

NS=$(kubectl get namespaces | awk '/wavefront-collector|wavefront|collector-targets|observability-system/ {print $1}')

if [[ ! -z "$NS" ]] ; then
    echo "Found Namespaces: ${NS}"
    kubectl delete --wait=false namespace ${NS} || true
fi


