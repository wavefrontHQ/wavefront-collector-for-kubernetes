#!/usr/bin/env bash

### LEGACY (MANUAL) ###

# collector DaemonSet logs
(kubectl --namespace wavefront-collector logs daemonset/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector DaemonSet found in wavefront-collector namespace; this is either a failure or it was installed some other way.") \
  > k8s-assist-legacy-wavefront-collector-legacy-daemonset.txt

# collector Deployment logs
(kubectl --namespace wavefront-collector logs deployment/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector Deployment found in wavefront-collector namespace; this is either a failure or it was installed some other way.") \
  > k8s-assist-legacy-wavefront-collector-legacy-deployment.txt

# Proxy Deployment logs
(kubectl --namespace default logs deployment/wavefront-proxy 2> /dev/null \
  || echo "No wavefront-proxy Deployment found in wavefront namespace; this may indicate a failure.") \
  > k8s-assist-legacy-wavefront-proxy.txt

### HELM ###

# collector DaemonSet logs
(kubectl --namespace wavefront logs daemonset/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector DaemonSet found in wavefront namespace; this is either a failure or it was installed some other way.") \
  > k8s-assist-helm-wavefront-collector-legacy-daemonset.txt

# Proxy Deployment logs
(kubectl --namespace wavefront logs deployment/wavefront-proxy 2> /dev/null \
  || echo "No wavefront-proxy Deployment found in wavefront namespace; this may indicate a failure or it was installed some other way.") \
  > k8s-assist-helm-wavefront-proxy.txt
  
### TMC ###

# collector DaemonSet logs
(kubectl --namespace tanzu-observability-saas logs daemonset/wavefront-collector 2> /dev/null \
  || echo "No wavefront-collector DaemonSet found in tanzu-observability-saas namespace; this is either a failure or it was installed some other way.") \
  > k8s-assist-helm-wavefront-collector-legacy-daemonset.txt

# Proxy Deployment logs
(kubectl --namespace tanzu-observability-saas logs deployment/wavefront-proxy 2> /dev/null \
  || echo "No wavefront-proxy Deployment found in tanzu-observability-saas namespace; this may indicate a failure or it was installed some other way.") \
  > k8s-assist-helm-wavefront-proxy.txt

### Operator ###

(kubectl --namespace observability-system get wavefront/wavefront 2> /dev/null \
  || echo "No wavefront resource found in observability-system namespace; this is either a failure or it was installed some other way.") \
  > k8s-assist-operator-status.txt

(kubectl  --namespace observability-system logs --selector='app.kubernetes.io/name=wavefront' 2> /dev/null \
    || echo "No operator logs found in observability-system namespace; this is either a failure or it was installed some other way.") \
    > k8s-assist-operator-logs.txt

zip k8s-assist-info.zip k8s-assist-*.txt
rm k8s-assist-*.txt
