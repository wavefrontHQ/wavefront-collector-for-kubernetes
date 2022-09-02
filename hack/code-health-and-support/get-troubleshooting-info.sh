#!/usr/bin/env bash

# legacy collector DaemonSet logs
(kubectl --namespace wavefront-collector logs daemonset/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector DaemonSet found in wavefront-collector namespace; this is either a failure or it was installed via the new Wavefront Operator.") \
  > k8s-assist-wavefront-collector-legacy-daemonset.txt

# legacy collector Deployment logs
(kubectl --namespace wavefront-collector logs deployment/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector Deployment found in wavefront-collector namespace; this is either a failure or it was installed via the new Wavefront Operator.") \
  > k8s-assist-wavefront-collector-legacy-deployment.txt

# Wavefront Proxy Deployment logs
(kubectl --namespace wavefront logs deployment/wavefront-proxy 2> /dev/null \
  || echo "No wavefront-proxy Deployment found in wavefront namespace; this may indicate a failure.") \
  > k8s-assist-wavefront-proxy.txt

# Operator collector DaemonSet logs
(kubectl --namespace wavefront logs daemonset/wavefront-node-collector 2> /dev/null \
  || echo "No wavefront-node-collector DaemonSet found in wavefront namespace; this is either a failure or it was installed via the legacy install method.") \
  > k8s-assist-wavefront-node-collector.txt

# Operator collector Deployment logs
(kubectl --namespace wavefront logs deployment/wavefront-cluster-collector 2> /dev/null \
  || echo "No wavefront-cluster-collector Deployment found in wavefront namespace; this is either a failure or it was installed via the legacy install method.") \
  > k8s-assist-wavefront-cluster-collector.txt

zip k8s-assist-info.zip k8s-assist-*.txt
rm k8s-assist-*.txt
