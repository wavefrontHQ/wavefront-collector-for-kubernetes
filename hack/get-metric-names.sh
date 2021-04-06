#!/bin/bash
set -e -u -o pipefail

NS=$(kubectl get namespaces | awk '/wavefront-collector/ {print $1}')
COLLECTOR=$(kubectl get pods -n $NS  | awk '/wavefront-collector-/ {print $1}')
echo "Gonna sleep for a bit..."
for i in `seq 90`; do
    echo -n "."
    sleep 10
done
echo "Ok, grabbing logs..."
kubectl logs $COLLECTOR -n $NS --since=5m | awk '/Metric/ {print $2}' >unsorted-metrics.csv
sort -u <unsorted-metrics.csv >sorted-metrics.csv
