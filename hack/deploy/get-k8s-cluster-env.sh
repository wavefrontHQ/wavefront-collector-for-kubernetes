#!/bin/bash -e

CURRENT_CONTEXT=$(kubectl config current-context)

if grep -q "kind" <<< "$CURRENT_CONTEXT"; then
  echo "Kind"
elif grep -q "gke" <<< "$CURRENT_CONTEXT"; then
  echo "GKE"
elif grep -q "eks" <<< "$CURRENT_CONTEXT"; then
  echo "EKS"
else
  echo "No matching env for ${CURRENT_CONTEXT}"
fi