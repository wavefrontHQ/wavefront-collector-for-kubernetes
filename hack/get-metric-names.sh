#!/bin/bash
set -e -u -o pipefail

NS=$(kubectl get namespaces | awk '/wavefront-collector/ {print $1}')
COLLECTOR=$(kubectl get pods -n $NS  | awk '/wavefront-collector-/ {print $1}')


PODS=`kubectl -n ${NS} get pod -l k8s-app=wavefront-collector | awk '{print $1}' | tail +2`

echo $PODS
#echo "Gonna sleep for a bit..."
#for i in `seq 30`; do
#    echo -n "."
#    sleep 10
#done

rm -f sorted-metrics.csv unsorted-metrics.csv
for pod in ${PODS} ; do
  echo "Ok, grabbing logs from pod ${pod}"
  kubectl logs $pod -n $NS --since=5m | awk '/Metric/ {print $2}' >>unsorted-metrics.csv
done

sort -u <unsorted-metrics.csv >sorted-metrics.csv
rm -f unsorted-metrics.csv


