# Metrics

## Kubernetes Source

Metrics collected per resource:

| Resource | Metrics |
|----------|---------|
| Cluster | CPU, Memory |
| Namespace | CPU, Memory |
| Nodes | CPU, Memory, Network, Filesystem, Storage, Uptime |
| Pods | CPU, Memory, Network, Filesystem, Storage, Uptime, Restarts |
| Pod_Containers | CPU, Memory, Filesystem, Storage, Accelerator, Uptime |
| System_Containers | CPU, Memory, Uptime |

Metrics collected per type:

| Metric Name | Description |
|------------|-------------|
| cpu.limit | CPU hard limit in millicores. |
| cpu.node_capacity | CPU capacity of a node. |
| cpu.node_allocatable | CPU allocatable of a node in millicores. |
| cpu.node_reservation | Share of CPU that is reserved on the node allocatable in millicores. |
| cpu.node_utilization | CPU utilization as a share of node allocatable in millicores. |
| cpu.request | CPU request (the guaranteed amount of resources) in millicores. |
| cpu.usage | Cumulative amount of consumed CPU time on all cores in nanoseconds. |
| cpu.usage_rate | CPU usage on all cores in millicores. |
| cpu.load | CPU load in milliloads, i.e., runnable threads * 1000. |
| memory.limit | Memory hard limit in bytes. |
| memory.major_page_faults | Number of major page faults. |
| memory.major_page_faults_rate | Number of major page faults per second. |
| memory.node_capacity | Memory capacity of a node. |
| memory.node_allocatable | Memory allocatable of a node. |
| memory.node_reservation | Share of memory that is reserved on the node allocatable. |
| memory.node_utilization | Memory utilization as a share of memory allocatable. |
| memory.page_faults | Number of page faults. |
| memory.page_faults_rate | Number of page faults per second. |
| memory.request | Memory request (the guaranteed amount of resources) in bytes. |
| memory.usage | Total memory usage. |
| memory.cache | Cache memory usage. |
| memory.rss | RSS memory usage. |
| memory.working_set | Total working set usage. Working set is the memory being used and not easily dropped by the kernel. |
| network.rx | Cumulative number of bytes received over the network. |
| network.rx_errors | Cumulative number of errors while receiving over the network. |
| network.rx_errors_rate | Number of errors while receiving over the network per second. |
| network.rx_rate | Number of bytes received over the network per second. |
| network.tx | Cumulative number of bytes sent over the network. |
| network.tx_errors | Cumulative number of errors while sending over the network. |
| network.tx_errors_rate | Number of errors while sending over the network. |
| network.tx_rate | Number of bytes sent over the network per second. |
| filesystem.usage | Total number of bytes consumed on a filesystem. |
| filesystem.limit | The total size of filesystem in bytes. |
| filesystem.available | The number of available bytes remaining in a the filesystem. |
| filesystem.inodes | The number of available inodes in a the filesystem. |
| filesystem.inodes_free | The number of free inodes remaining in a the filesystem. |
| ephemeral_storage.limit | Local ephemeral storage hard limit in bytes. |
| ephemeral_storage.request | Local ephemeral storage request (the guaranteed amount of resources) in bytes. |
| ephemeral_storage.usage | Total local ephemeral storage usage. |
| ephemeral_storage.node_capacity | Local ephemeral storage capacity of a node. |
| ephemeral_storage.node_allocatable | Local ephemeral storage allocatable of a node. |
| ephemeral_storage.node_reservation | Share of local ephemeral storage that is reserved on the node allocatable. |
| ephemeral_storage.node_utilization | Local ephemeral utilization as a share of ephemeral storage allocatable. |
| accelerator.memory_total | Memory capacity of an accelerator. |
| accelerator.memory_used | Memory used of an accelerator. |
| accelerator.duty_cycle | Duty cycle of an accelerator. |
| accelerator.request | Number of accelerator devices requested by container. |
| uptime  | Number of milliseconds since the container was started. |

## Kubernetes State Source
| Resource | Metric Name | Description |
|----------|---------|-------------|
| Deployment | deployment.desired_replicas | Number of desired pods. |
| Deployment | deployment.available_replicas | Total number of available pods (ready for at least minReadySeconds). |
| Deployment | deployment.ready_replicas | Total number of ready pods. |
| Replicaset | replicaset.desired_replicas | Number of desired replicas. |
| Replicaset | replicaset.available_replicas | Number of available replicas (ready for at least minReadySeconds). |
| Replicaset | replicaset.ready_replicas | Number of ready replicas. |
| Daemonset | daemonset.desired_scheduled | Total number of nodes that should be running the daemon pod. |
| Daemonset | daemonset.current_scheduled | Number of nodes that are running at least 1 daemon pod and are supposed to run the daemon pod. |
| Daemonset | daemonset.misscheduled | Number of nodes that are running the daemon pod, but are not supposed to run the daemon pod. |
| Daemonset | daemonset.ready | Number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready. |
| Statefulset | statefulset.desired_replicas | Number of desired replicas. |
| Statefulset | statefulset.current_replicas | Number of Pods created by the StatefulSet controller from the StatefulSet version indicated by currentRevision.
| Statefulset | statefulset.ready_replicas | Number of Pods created by the StatefulSet controller that have a Ready Condition. |
| Statefulset | statefulset.updated_replicas | Number of Pods created by the StatefulSet controller from the StatefulSet version indicated by updateRevision. |
| Job | job.active | Number of actively running pods. |
| Job | job.failed | Number of pods which reached phase Failed. |
| Job | job.succeeded | Number of pods which reached phase Succeeded. |
| Job | job.completions | Desired number of successfully finished pods the job should be run with. -1.0 indicates the value was not set. |
| Job | job.parallelism | Maximum desired number of pods the job should run at any given time. -1.0 indicates the value was not set. |
| CronJob | cronjob.active | Number of currently running jobs. |
| HorizontalPodAutoscaler | hpa.desired_replicas | Desired number of replicas of pods managed by this autoscaler as last calculated by the autoscaler. |
| HorizontalPodAutoscaler | hpa.min_replicas | Lower limit for the number of replicas to which the autoscaler can scale down. |
| HorizontalPodAutoscaler | hpa.max_replicas | Upper limit for the number of replicas to which the autoscaler can scale up. |
| HorizontalPodAutoscaler | hpa.current_replicas | Current number of replicas of pods managed by this autoscaler, as last seen by the autoscaler. |

## Prometheus Source
Varies by scrape target.

## Systemd Source

| Metric Name | Description |
|------------|-------------|
| kubernetes.systemd.unit.state | Unit state (active, inactive etc). |
| kubernetes.systemd.unit.start.time.seconds | Start time of the unit since epoch in seconds. |
| kubernetes.systemd.system.running | Whether the system is operational (`systemctl is-system-running`). |
| kubernetes.systemd.units | Top level summary of systemd unit states (# of active, inactive units etc). |
| kubernetes.systemd.service.restart.total | Service unit count of Restart triggers. |
| kubernetes.systemd.timer.last.trigger.seconds | Seconds since epoch of last trigger. |
| kubernetes.systemd.socket.accepted.connections.total | Total number of accepted socket connections. |
| kubernetes.systemd.socket.current.connections | Current number of socket connections. |
| kubernetes.systemd_socket_refused_connections_total | Total number of refused socket connections. |

## Telegraf Source

| Metric Prefix | Metrics Collected |
|------------|-------------|
| mem. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mem#metrics) |
| net. | [metrics list](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/net/NET_README.md#measurements--fields) |
| netstat. | [metrics list](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/net/NETSTAT_README.md#measurements) |
| linux.sysctl.fs. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/linux_sysctl_fs#linux-sysctl-fs-input) |
| swap. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/swap#metrics) |
| cpu. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/cpu#measurements) |
| disk. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/disk#metrics) |
| diskio. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/diskio#metrics) |
| system. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/system#metrics) |
| kernel. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/kernel#measurements--fields) |
| processes. | [metrics list](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/processes#measurements--fields) |

## Collector Health Metrics

| Metric Name | Description |
|------------|-------------|
| kubernetes.collector.discovery.enabled | Whether discovery is enabled. 0 (false) or 1 (true). |
| kubernetes.collector.discovery.rules.count | # of discovery configuration rules. |
| kubernetes.collector.discovery.targets.registered | # of auto discovered scrape targets currently being monitored. |
| kubernetes.collector.events.* | Events received, sent and filtered. |
| kubernetes.collector.leaderelection.error | leader election error counter. Only emitted in daemonset mode. |
| kubernetes.collector.leaderelection.leading | 1 indicates a pod is the leader. 0 (no). Only emitted in daemonset mode. |
| kubernetes.collector.runtime.* | Go runtime metrics (MemStats, NumGoroutine etc). |
| kubernetes.collector.sink.manager.timeouts | Counter of timeouts in sending data to Wavefront. |
| kubernetes.collector.source.manager.providers | # of configured source providers. Includes sources configured via auto-discovery. |
| kubernetes.collector.source.manager.scrape.errors | Scrape error counter across all sources. |
| kubernetes.collector.source.manager.scrape.latency.* | Scrape latencies across all sources. |
| kubernetes.collector.source.manager.scrape.timeouts | Scrape timeout counter across all sources. |
| kubernetes.collector.source.manager.sources | # of configured scrape targets. For example, a single Kubernetes source provider on a 10 node cluster will yield a count of 10. |
| kubernetes.collector.source.points.collected | collected points counter per source type. |
| kubernetes.collector.source.points.filtered | filtered points counter per source type. |
| kubernetes.collector.version | The version of the collector. |
| kubernetes.collector.wavefront.points.* | Wavefront sink points sent, filtered, errors etc. |
| kubernetes.collector.wavefront.events.* | Wavefront sink events sent, filtered, errors etc. |
| kubernetes.collector.wavefront.sender.type | 1 for proxy and 0 for direct ingestion. |
