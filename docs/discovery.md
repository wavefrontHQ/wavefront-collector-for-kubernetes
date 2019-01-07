# Auto Discovery

The Wavefront Kubernetes Collector can auto discover pods and services that expose prometheus format metrics and dynamically configure [Prometheus scrape targets](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/configuration.md#prometheus-source) for the targets.

Pods/Services are discovered based on annotations and based on discovery rules provided using a configuration file.

## Annotations Based Discovery
Pods/Services annotated with `prometheus.io/scrape` set to **true** will be auto discovered.

Additional annotations that apply:
- `prometheus.io/scheme`: Defaults to **http**.
- `prometheus.io/path`: Defaults to **/metrics**.
- `prometheus.io/port`: Defaults to a port free target if omitted.
- `prometheus.io/prefix`: Dot suffixed string to prefix reported metrics. Defaults to an empty string.
- `prometheus.io/includeLabels`: Whether to include pod labels as tags on reported metrics. Defaults to **true**.
- `prometheus.io/source`: Optional source for the reported metrics. Defaults to **prom_source**.

## Rules Based Discovery
Discovery rules enable discovery based on labels and namespaces. Prometheus scrape options similar to the annotations above are supported.

The rules are provided to the collector using the optional `--discovery-config` flag. When provided, the collector watches for configuration changes and automatically reloads configurations without having to restart the collector.

The sample configuration below enables discovery of `kube-dns` pods and a `my-app` application pods:
```yaml
global:
  discovery_interval: 10m
prom_configs:
- name: kube-dns
  labels:
    k8s-app: kube-dns
  namespace: kube-system
  port: 10054
  prefix: kube.dns.
- name: my-app
  labels:
    app: my-app
    service: ingestion
  namespace: my-app-namespace
  prefix: my-app.
```

## Use Cases
Together, annotation and rule based discovery can be used to easily collect metrics from the Kubernetes control plane (kube-dns etc), NGINX ingresses, and any application that exposes a Prometheus scrape endpoint.

## Disabling Auto Discovery
Auto discovery is enabled by default and can be disabled by setting the `--enable-discovery` collector flag to `false`.
