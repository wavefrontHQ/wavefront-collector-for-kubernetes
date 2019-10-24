# Installation and configuration of Wavefront Collector Operator on OpenShift
This page contains the Installation and Configuration steps to monitor Openshift using Wavefront Collector Operator.

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
4. Create Wavefront Collector Operator CRD:
```
 oc create -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/crds/wavefront-collector-operator_v1beta1_crd.yaml
```
5. Install the operator:
```
oc create -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/operator.yaml
```
6. Download the Custom Resource (CR) YAML file that can be customized and used as a blueprint to deploy the collector.
```
curl -O https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/crds/wavefront-collector-operator_v1beta1_cr.yaml 
```

7. Configure Wavefront Proxy:

* Set `collector.useProxy: true` in `wavefront-collector-operator_v1beta1_cr.yaml`

Now configure the Wavefront proxy by following any one of the options:

* Option 1: [Using external Wavefront proxy](#option-1-using-external-wavefront-proxy) to send metrics to a Wavefront proxy that is already configured and running outside of the cluster.
* Option 2: [Using internal Wavefront proxy](#option-2-using-internal-wavefront-proxy) to deploy Wavefront proxy into the Openshift cluster and configuring it.

8. Replace OPENSHIFT_CLUSTER_NAME in `wavefront-collector-operator_v1beta1_cr.yaml` and run:
```
oc create -f wavefront-collector-operator_v1beta1_cr.yaml -n wavefront-collector
``` 


### Option 1. Using external Wavefront proxy 

* Uncomment the property `collector.proxyAddress` in `wavefront-collector-operator_v1beta1_cr.yaml`and provide the external Wavefront proxy IP address with port.

### Option 2. Using internal Wavefront proxy

* Create a file `pvc.yaml` with below content and replace `<storage_class_name>` with available storage class in the cluster.

```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: wavefront-storage
  namespace: wavefront-collector
  annotation:
    <storage_class_name>
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

* Now create storage `wavefront-storage`
```
  oc create -f pvc.yaml
```

* Set `proxy.enabled: true` and replace `YOUR_CLUSTER_NAME`, `YOUR_API_TOKEN` and `WF_PROXY_STORAGE` in `wavefront-collector-operator_v1beta1_cr.yaml`


## Uninstalling the Operator and its components

```
oc delete -f wavefront-collector-operator_v1beta1_cr.yaml -n wavefront-collector
oc delete -f pvc.xml -n wavefront-collector
oc delete -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/operator.yaml
oc delete -f https://raw.githubusercontent.com/wavefrontHQ/wavefront-kubernetes-collector/master/deploy/openshift/openshift-operator/deploy/crds/wavefront-collector-operator_v1beta1_crd.yaml

```

