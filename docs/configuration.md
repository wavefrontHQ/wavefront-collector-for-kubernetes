# Configuration

This page documents advanced configuration options for various aspects of the Wavefront Kubernetes Collector.

## Kubernetes Source
- `kubeletPort`: Defaults to 10255. Use 10250 for the secure port.
- `kubeletHttps`: Defaults to false. Set to true if `kubeletPort` set to 10250.
- `inClusterConfig`: Defaults to true.
- `useServiceAccount`: Defaults to false.
- `auth`: If using secure kubelet port, this can be set to a valid kubeConfig file provided using a config map.

Example Usage:
```
--source=kubernetes.summary_api:https://kubernetes.default.svc?kubeletPort=10250&kubeletHttps=true&inClusterConfig=false&auth=/etc/kubernetes/kubeconfig.conf
```

See [configs.go](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/internal/kubernetes/configs.go) for how these properties are used.

## Prometheus Source
- `url`: The URL for a Prometheus metrics endpoint. Service URLs work across namespaces.
- `prefix`: The prefix (dot suffixed such as `prom.`) to be applied to all metrics for this source. Defaults to empty string.
- `source`: The source to set for the metrics from this source. Defaults to `prom_source`.

Example Usage:
```
--source=prometheus:''?url=http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics&prefix=prom.```

## Wavefront Sink
- `server`: The Wavefront URL of the form `https://YOUR_INSTANCE.wavefront.com`. Only required for direct ingestion.
- `token`: The Wavefront API token with direct data ingestion permission. Only required for direct ingestion.
- `proxyAddress`: The Wavefront proxy service address of the form `wavefront-proxy.default.svc.cluster.local:2878`. Requires the proxy to be deployed in Kubernetes.
- `clusterName`: A unique identifier for your Kubernetes cluster. Defaults to `k8s-cluster`. This is included as a point tag on all metrics sent to Wavefront.
- `includeLabels`: If set to true, any Kubernetes labels will be applied to metrics as tags. Defaults to false.
- `includeContainers`: If set to true, all container metrics will be sent to Wavefront. When set to false, container level metrics are skipped (pod level and above are still sent to Wavefront). Defaults to true.
- `prefix`: The global prefix (dot suffixed) to be added for Kubernetes metrics. Defaults to `kubernetes.`. This does not apply to other sources. Use source level prefixes for sources other than the `kubernetes` source.

Example Usages:
```
# Direct Ingestion
--sink=wavefront:?server=https://YOUR_INSTANCE.wavefront.com&token=YOUR_TOKEN&clusterName=k8s-cluster&includeLabels=true

# Proxy
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true
```
