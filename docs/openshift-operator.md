# Installation and configuration of Operator on OpenShift
This page contains the Installation and Configuration steps to monitor Openshift using Wavefront Kubernetes Collector Operator.

**Supported Versions**: Openshift Enterprise 3.11

1. Log in to the Openshift master node.
2. Log in to the Openshift cluster:
```
oc login -u <ADMIN_USER>
```
3. Create `wavefront-collector` project:

```
 oc adm new-project --node-selector="" wavefront-collector
```
4. Create Wavefront collector operator CRD:
```
 oc create -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/crds/wavefront-collector-operator_v1beta1_crd.yaml
```
5. Install the operator:
```
oc create -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/operator.yaml
```
6. Download the Custom Resource (CR) YAML file that can be customized and used as a blueprint to deploy the collector
```
curl -O https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/crds/wavefront-collector-operator_v1beta1_cr.yaml 
```

7. Replace OPENSHIFT_CLUSTER_NAME, YOUR_CLUSTER_NAME and YOUR_API_TOKEN in `wavefront-collector-operator_v1beta1_cr.yaml` and run:
```
oc create -f wavefront-collector-operator_v1beta1_cr.yaml
``` 
