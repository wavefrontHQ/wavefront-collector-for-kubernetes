# Auto Discovery

The Wavefront Collector for Kubernetes can auto-discover pods and services that expose metrics, and dynamically start collecting metrics for the targets.

Pods/Services can be discovered based on annotations and discovery rules. Discovery rules are provided via the configuration file.

## Table of Contents
* [Annotation based discovery](#annotation-based-discovery)
* [Rule based discovery](#rule-based-discovery)
* [Use Cases](#use-cases)
* [Disabling Discovery](#disabling-auto-discovery)

## Annotation based discovery
**Note**: Annotation based discovery is only supported for prometheus endpoints currently.

[Annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) are metadata you attach to Kubernetes objects. Amongst other uses, they can act as pointers for monitoring tools.

The collector can dynamically discover pods/services annotated with `prometheus.io/scrape` set to **true**. Additional annotations can be provided to inform the collector on how to perform the collection and what prefix and tags to add to the emitted metrics.

Additional annotations that apply:
- `prometheus.io/scheme`: Defaults to **http**.
- `prometheus.io/path`: Defaults to **/metrics**.
- `prometheus.io/port`: Defaults to a port free target if omitted.
- `prometheus.io/prefix`: Dot suffixed string to prefix reported metrics. Defaults to an empty string.
- `prometheus.io/includeLabels`: Whether to include Kubernetes labels as tags on reported metrics. Defaults to **true**.
- `prometheus.io/source`: Optional source for the reported metrics. Defaults to the node name on which collection is performed.
- `prometheus.io/collectionInterval`: Custom collection interval. Defaults to 1m. Format is `[0-9]+(ms|[smhdwy])`.
- `prometheus.io/insecureSkipVerify`: Whether to skip https cert validation. Defaults to true.
- `prometheus.io/serverName`: The cert hostname to verify for the discovered targets.

See an [example](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/deploy/examples/prometheus-annotations-example.yaml) for how to annotate a pod with the above annotations.

### Disabling annotation discovery
Discovery based on annotations is enabled by default, but can be disabled by setting the `disable_annotation_discovery` configuration option to `true`:

```
discovery:
  disable_annotation_discovery: true
```

## Rule based discovery
Discovery rules encompass a few distinct aspects:
- *Selectors*: The criteria for identifying matching kubernetes resources using container images, resource labels and/or namespaces.
- *Plugin Type*: The [type](#plugin-types) of source plugin to use for collecting metrics from the discovered targets.
- *Plugin Config*: Configuration information on how to collect metrics from the discovered targets.
- *Transformations*: Adds prefix, tags and filters on the collected data before emitting them to Wavefront.

The rules can be supplied under the `discovery` section within the top-level `--config-file` or dynamically via [runtime configurations](#runtime-configurations).

The collector fetches all the pods/services on startup. It also listens for runtime changes. The rules that match a pod/service are used to collect metrics from the matching targets.

### Configuration file
Source: [configs.go](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/internal/discovery/configs.go)

The configuration file is YAML based. Each rule has the following structure:
```yaml
# Unique name per rule. Used internally as map keys and thus needs to be unique per rule.
name: <string>

# Plugin type to use for collecting metrics. Example: 'prometheus' or 'telegraf/redis'
type: <string>

# Selectors for identifying matching kubernetes resources.
# One of images, labels or namespaces is required.
selectors:
  # pod | service. Defaults to pod.
  resourceType: <string>

  # The container images to match against. Provided as a list of glob pattern strings. Ex: 'redis*'
  images:
    - 'redis:*'
    - '*redis*'

  # map of labels to select resources by. Label values are provided as a list of glob pattern strings.
  labels:
    k8s-app:
    - 'redis'
    - '*cache*'

  # namespaces to filter resources by. Provided as a list of glob pattern strings.
  namespaces:
  - default

# The port to be monitored on the pod or service
port: <string>

# The scheme to use. Defaults to "http".
scheme: <string>

# Defaults to "/metrics" for prometheus plugin type. Empty string for telegraf plugins.
path: <string>

# The configuration specific to a plugin.
# For telegraf based plugins config is provided in toml format: https://github.com/toml-lang/toml
# and parsed using https://github.com/influxdata/toml
conf: <multi_line_string>

# Optional static source for metrics collected using this rule. Defaults to agent node name.
source: <string>

# Optional prefix for metrics collected using this rule. Defaults to empty string.
prefix: <string>

# Optional map of custom tags to include with the reported metrics
tags: <map of key-value pairs>

# Whether to include resource labels with the reported metrics. Defaults to "true".
includeLabels: <true|false>

# filters applied towards the collected metrics before emitting them.
filters:
  # see the filtering documentation:
  # https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/filtering.md

# custom collection interval for this rule
collection:
  # Duration type specified as [0-9]+(ms|[smhdwy])
  interval: 30s

  # Duration type specified as [0-9]+(ms|[smhdwy])
  timeout: 20s
```
See the reference [example](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/deploy/examples/conf.example.yaml) for details on how to specify the discovery rules.

### Plugin Types
The supported plugin types are:
- **prometheus**: For collecting metrics from prometheus metric endpoints.
- **telegraf/pluginName**: For collecting metrics from applications that are supported by telegraf. See [here](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/blob/master/docs/metrics.md#telegraf-source) for the list of supported applications.

  **Note:** The version of telegraf embedded within the collector is 1.10.x.

### Runtime Configurations
Runtime configurations allow specifying discovery rules via [configmaps](https://kubernetes.io/docs/concepts/configuration/configmap/) or [secrets](https://kubernetes.io/docs/concepts/configuration/secret/) outside of the main configuration file.

This feature can be enabled in the main config file:
```yaml
discovery:
  # flag to enable runtime configurations
  enable_runtime_plugins: true

  # frequency of evaluating changes to runtime configs (adds/updates/deletes)
  discovery_interval: 5m
```
The runtime configmaps should be annotated with `wavefront.com/discovery-config: 'true'` and deployed under the same namespace as the Wavefront collector.

Discovery rules specified in the main config and via runtime configs are combined together to form a single set of rules to drive auto-discovery decisions.

The `discovery_interval` controls how often runtime config changes are evaluated. This is pertinent as runtime changes requires the collector to re-evaluate all pods/services for discovery/data collection.

See the reference [configmap example](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/deploy/examples/runtime/memcached-runtime-config.yaml) or [secret example](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/deploy/examples/runtime/memcached-runtime-secret-config.yaml) for details.

## Use Cases
Together, annotation and rule based discovery can be used to easily collect metrics from the Kubernetes control plane (apiserver, etcd, dns etc), NGINX ingresses, and any application that exposes a Prometheus scrape endpoint.

## Disabling Auto Discovery
Auto discovery is enabled by default and can be disabled by setting the `enableDiscovery` configuration option to `false`.
