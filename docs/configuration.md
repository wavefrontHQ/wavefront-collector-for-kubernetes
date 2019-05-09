# Configuration

This page documents advanced configuration options for various aspects of the Wavefront Kubernetes Collector.

## Wavefront Collector
```
Usage of ./wavefront-collector:
      --alsologtostderr                     log to standard error as well as files
      --discovery-config string             optional discovery configuration file
      --enable-discovery                    enable pod discovery (default true)
      --ignore-label strings                ignore this label when joining labels
      --label-separator string              separator used for joining labels (default ",")
      --log-backtrace-at traceLocation      when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                      If non-empty, write log files in this directory
      --log-flush-frequency duration        Maximum number of seconds between log flushes (default 5s)
      --logtostderr                         log to standard error instead of files (default true)
      --max-procs int                       max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)
      --metric-resolution duration          The resolution at which the collector will retain metrics. (default 1m0s)
      --sink *flags.Uris                    external sink(s) that receive data (default [])
      --sink-export-data-timeout duration   Timeout for exporting data to a sink (default 20s)
      --source *flags.Uris                  source(s) to watch (default [])
      --stderrthreshold severity            logs at or above this threshold go to stderr (default 2)
      --store-label strings                 store this label separately from joined labels with the same name (name) or with different name (newName=name)
  -v, --v Level                             log level for V logs
      --version                             print version info and exit
      --vmodule moduleSpec                  comma-separated list of pattern=N settings for file-filtered logging
```

## Kubernetes Source
- `kubeletPort`: Defaults to 10255. Use 10250 for the secure port.
- `kubeletHttps`: Defaults to false. Set to true if `kubeletPort` set to 10250.
- `inClusterConfig`: Defaults to true.
- `useServiceAccount`: Defaults to false.
- `auth`: If using secure kubelet port, this can be set to a valid kubeConfig file provided using a config map.

Example usage using secure port and service account:
```
--source=kubernetes.summary_api:https://kubernetes.default.svc?useServiceAccount=true&kubeletHttps=true&kubeletPort=10250&insecure=true
```

See [configs.go](https://github.com/wavefronthq/wavefront-kubernetes-collector/tree/master/internal/kubernetes/configs.go) for how these properties are used.

## Prometheus Source
- `url`: The URL for a Prometheus metrics endpoint. Kubernetes Service URLs work across namespaces.
- `prefix`: The prefix (dot suffixed such as `prom.`) to be applied to all metrics for this source. Defaults to empty string.
- `source`: The source to set for the metrics from this source. Defaults to `prom_source`.
- `tag`: Custom tags to include with metrics reported by this source, of the form `tag=key1:val1&tag=key2:val2`.

Example Usage:
```
--source=prometheus:''?url=http://kube-state-metrics.kube-system.svc.cluster.local:8080/metrics&prefix=prom.
```

## Telegraf Source
- `prefix`: The prefix to be applied to all metrics for this source. Defaults to empty string.
- `plugins`: Comma separated list of telegraf plugins to collect metrics from. Defaults to collecting from all plugins.

The list of plugins that are supported:
- mem, net, netstat, linux_sysctl_fs, swap, cpu, disk, diskio, system, kernel, processes.

Example Usage:
```
--source=telegraf:''?prefix=telegraf.&plugins=cpu,netstat,disk,diskio
```

## Systemd Source
- `prefix`: The prefix to be applied to all metrics for this source. Defaults to `kubernetes.systemd.`.
- `taskMetrics`: Defaults to true. Set to false to not collect systemd unit task metrics.
- `startTimeMetrics`: Defaults to true. Set to false to not collect system unit start time metrics.
- `restartMetrics`: Defaults to false. Set to true to collect systemd unit restart metrics.
- `unitWhitelist`: List of glob patterns. Only unit names matching the whitelist are monitored. Defaults to all units.
- `unitBlacklist`: List of glob patterns. Unit names matching the blacklist are not monitored. Defaults to empty string.

Example Usage:
```
--source=systemd:''?prefix=kubernetes.systemd.&restartMetrics=true&unitWhitelist=*docker*&unitWhitelist=*kubelet*
```

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
## Direct Ingestion
--sink=wavefront:?server=https://YOUR_INSTANCE.wavefront.com&token=YOUR_TOKEN&clusterName=k8s-cluster&includeLabels=true

## Proxy
--sink=wavefront:?proxyAddress=wavefront-proxy.default.svc.cluster.local:2878&clusterName=k8s-cluster&includeLabels=true
```
