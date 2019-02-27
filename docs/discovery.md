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
- `prometheus.io/source`: Optional source for the reported metrics. Defaults to the name of the Kubernetes resource.

## Rules Based Discovery
Discovery rules enable discovery based on labels and namespaces. Prometheus scrape options similar to the annotations above are supported.

The rules are provided to the collector using the optional `--discovery-config` flag. When provided, the collector watches for configuration changes and automatically reloads configurations without having to restart it.

### Configuration file
The configuration file is YAML based and has the following structure:
```yaml
global:
  # Frequency of rule based target discovery
  discovery_interval: 10m

# List of rules for auto discovering Prometheus scrape targets
prom_configs
```
The structure for the `prom_config`:
```yaml
# Name describing the rule
name: rule_name
# The resource type the rule applies to. Defaults to pod.
resourceType: <pod|service|apiserver>
# Map of kubernetes labels identifying the pods and services. Does not apply for apiserver.
labels:
  <key1>: <val1>
  <key2>: <val2>
# Optional namespace to filter by for pods and services.
namespace: <my-app-namespace>
# Optional port to scrape for Prometheus metrics. Defaults to a port-free target.
port: <port_number>
# Optional scheme to use. Defaults to http.
scheme: <http|https>
# Optional dot suffixed prefix to apply to metrics collected using this rule.
prefix: <some.prefix.>
# Optional map of custom tags to include with the metrics collected using this rule.
tags:
  <key1>:<val1>
  <key2>:<val2>
# Optional source for metrics collected using this rule. Defaults to the name of the Kubernetes resource.
source: <source_name>
# Whether to include Kubernetes resource labels with the reported metrics. Defaults to "true".
includeLabels: <true|false>
# Optional filtering rules to apply towards metrics collected using this rule.
filters:
  # See filtering documentation
```
See the [filtering documentation](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/filtering.md) for details on filtering the metrics that are reported to Wavefront.

### Sample Configuration file
The sample configuration below enables discovery of the `apiserver`, `kube-dns` and a `my-app` application pods:
```yaml
global:
  discovery_interval: 10m
prom_configs:
- name: kube-apiserver
  resourceType: apiserver
  scheme: https
  port: 443
  prefix: kube.apiserver.
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
See the [sample deployment](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/tree/master/deploy/discovery-examples) for details on how to deploy the discovery rules.

## Use Cases
Together, annotation and rule based discovery can be used to easily collect metrics from the Kubernetes control plane (apiserver, etcd, dns etc), NGINX ingresses, and any application that exposes a Prometheus scrape endpoint.

## Disabling Auto Discovery
Auto discovery is enabled by default and can be disabled by setting the `--enable-discovery` collector flag to `false`.
