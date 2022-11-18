#!/bin/bash -e

if ! command -v kubectl &> /dev/null
then
    echo "No context"
    exit 0
fi

CURRENT_CONTEXT=$(kubectl config current-context)

if grep -q "kind" <<< "$CURRENT_CONTEXT"; then
  echo "Kind"
elif grep -q "gke" <<< "$CURRENT_CONTEXT"; then
  echo "GKE"
elif grep -q "aks" <<< "$CURRENT_CONTEXT"; then
  echo "AKS"
elif grep -q "k8po-ci" <<< "$CURRENT_CONTEXT"; then
  echo "AKS"
elif grep -q "eks" <<< "$CURRENT_CONTEXT"; then
  echo "EKS"
else
  echo "No matching env for ${CURRENT_CONTEXT}"
fi