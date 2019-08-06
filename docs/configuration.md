# Configuration

The Wavefront Kubernetes Collector is configured via command-line flags and a configuration file.

Starting with version 1.0, most command line flags have been deprecated in favor of a top-level configuration file.

## Flags
```
Usage of ./wavefront-collector:
      --config-file string             required configuration file
      --daemon                         enable daemon mode (required when running as daemonset)
      --log-level string               one of info, debug or trace (default "info")
      --profile                        enable pprof (for debugging)
      --version                        print version info and exit
```

## Configuration file

Source: [config.go](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/internal/configuration/config.go)

The configuration file is written in YAML and provided using the `--config-file` flag. The Collector can reload configuration changes at runtime.

A reference example is provided [here](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/deploy/examples/conf.example.yaml).

```yaml
# An unique identifier for your Kubernetes cluster. Defaults to 'k8s-cluster'.
# Included as a point tag on all metrics reported to Wavefront.
clusterName: k8s-cluster

# Whether auto-discovery is enabled. Defaults to true.
enableDiscovery: true

# The global interval at which data is flushsed. Defaults to 60 seconds.
# Duration type specified as [0-9]+(ms|[smhdwy])
flushInterval: 60s

# The global interval at which data is collected. Defaults to 60 seconds.
# Duration type specified as [0-9]+(ms|[smhdwy])
# Note: collection intervals can be overridden per source.
defaultCollectionInterval: 60s

# Timeout for sinks to export data to Wavefront. Defaults to 20 seconds.
# Duration type specified as [0-9]+(ms|[smhdwy])
sinkExportDataTimeout: 20s

# runtime.GOMAXPROCS value to use.
maxProcs: 4

# Required: List of Wavefront sinks. At least 1 required.
sinks:
  # see the Wavefront sink section for details

sources:
  # Required: Source for collecting metrics from the stats summary API.
  kubernetes_source:
    # see kubernetes_source for details

  # Optional source for emitting internal collector stats.
  internal_stats_source:
    # see internal_stats_source for details

  # Optional list of prometheus sources.
  prometheus_sources:
    # see prometheus_source for details

  # Optional list of telegraf sources.
  telegraf_sources:
    # see telegraf_source for details

  # Optional source for collecting host level systemd unit metrics.
  systemd_source:
    # see systemd_source for details

# Optional list of auto-discovery rules.
discovery_configs:
  # see auto-discovery for details
```

### Wavefront sink

```yaml
# The Wavefront proxy address of the form 'hostname:port'.
proxyAddress: wavefront-proxy.default.svc.cluster.local:2878

# Wavefront URL of the form https:YOUR_INSTANCE.wavefront.com. Only required for direct ingestion.
server: https://<instance>.wavefront.com

# Wavefront API token with direct data ingestion permission. Only required for direct ingestion.
token: <string>
```


### kubernetes_source

```yaml
# Defaults to empty string when using port 10255.
url: 'https://kubernetes.default.svc'

# Either 10255 (default, read-only kubelet port) or 10250 (secure kubelet port).
kubeletPort: <10250|10255>

# Defaults to false. Set to true if `kubeletPort` set to 10250.
kubeletHttps: <true|false>

# Defaults to true.
inClusterConfig: <true|false>

# Defaults to false.
useServiceAccount: <true|false>

# Defaults to false.
insecure: <true|false>

# Optional: a valid kubeConfig file provided using a config map
auth: <string>
```

See [configs.go](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/internal/kubernetes/configs.go) for how these properties are used.

### prometheus_source

```yaml
# The URL for a prometheus metrics endpoint. Kubernetes service URLs work across namespaces.
url: <string>

# Optional HTTP configuration
httpConfig:
  [ <ClientConfig> ]

# The source (tag) to set for the metrics collected by this source. Defaults to node name.
source: <string>
```

### telegraf_source
```yaml
# The list of plugins to be enabled. Empty list defaults to enabling all plugins.
# Supported plugins are: mem, net, netstat, linux_sysctl_fs, swap, cpu, disk, diskio, system, kernel, processes
plugins: []
```

### systemd_source
```yaml
# Whether to include systemd task metrics. Defaults to true.
taskMetrics: <true|false>

# Whether to include systemd start time metrics. Defaults to true.
startTimeMetrics: <true|false>

# Whether to include restart metrics. Defaults to false.
restartMetrics: <true|false>

# List of glob patterns. Metrics from matching systemd unit names are reported.
unitWhitelist:
- 'docker*'
- 'kubelet*'

# List of glob patterns. Metrics from matching systemd unit names are not reported.
unitBlacklist:
- '*mount*'
- 'etc*'
```

### Common properties
#### Prefix, tags and filters
All sources and sinks support the following common properties:
```yaml
# An optional dot suffixed prefix for metrics emitted by the sink or source.
prefix: <string>

# A map of key value pairs that are included as point tags on all metrics emitted
# by the sink or source.
tags:
  env: non-production
  region: us-west-2

# Filters to be applied to metrics collected by a source or reported by sinks.
filters:
  # List of glob patterns. Only metrics with names matching the whitelist are reported.
  metricWhitelist:
  - 'kube.dns.http.*'
  - 'kube.dns.process.*'

  # List of glob patterns. Metrics with names matching the blacklist are dropped.
  metricBlacklist:
  - 'kube.dns.go.*'

  # Map of tag names to list of glob patterns. Only metrics containing tag keys and values matching the whitelist will be reported.
  metricTagWhitelist:
    env: 'prod*'

  # Map of tag names to list of glob patterns. Metrics containing blacklisted tag keys and values will be dropped.
  metricTagBlacklist:
    env: 'test*'

  # List of glob patterns. Tags with matching keys will be included. All other tags will be excluded.
  tagInclude:
  - namespace
  - 'label.app'
  - 'label.component'

  # List of glob patterns. Tags with matching keys will be excluded.
  tagExclude:
  - handler
  - image
```
#### Custom collection intervals
All sources support using a custom collection interval:
```yaml
# Duration type specified as [0-9]+(ms|[smhdwy])
interval: 30s

# Duration type specified as [0-9]+(ms|[smhdwy])
timeout: 20s
```
