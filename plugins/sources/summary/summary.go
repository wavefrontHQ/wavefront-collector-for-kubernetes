// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/sources/summary/summary.go
// Diff against master for changes to the original code.

// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"fmt"
	"net"
	"time"

	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	. "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"

	kube_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

// Prefix used for the LabelResourceID for volume metrics.
const VolumeResourcePrefix = "Volume:"

var collectErrors gm.Counter

func init() {
	pt := map[string]string{"type": "kubernetes.summary_api"}
	collectErrors = gm.GetOrRegisterCounter(reporting.EncodeKey("source.collect.errors", pt), gm.DefaultRegistry)
}

type NodeInfo struct {
	NodeName       string
	HostName       string
	HostID         string
	KubeletVersion string
	NodeRole       string
	IP             net.IP
}

// Kubelet-provided metrics for pod and system container.
type summaryMetricsSource struct {
	node          NodeInfo
	kubeletClient *kubelet.KubeletClient
}

func NewSummaryMetricsSource(node NodeInfo, client *kubelet.KubeletClient) Source {
	return &summaryMetricsSource{
		node:          node,
		kubeletClient: client,
	}
}

func (src *summaryMetricsSource) AutoDiscovered() bool {
	return false
}

func (src *summaryMetricsSource) Name() string {
	return src.String()
}

func (src *summaryMetricsSource) Cleanup() {}

func (src *summaryMetricsSource) String() string {
	return fmt.Sprintf("kubelet_summary:%s:%d", src.node.IP, src.kubeletClient.GetPort())
}

func (src *summaryMetricsSource) Scrape() (*Batch, error) {
	result := &Batch{
		Timestamp: time.Now(),
		Sets:      map[metrics.ResourceKey]*Set{},
	}

	summary, err := func() (*stats.Summary, error) {
		return src.kubeletClient.GetSummary(src.node.IP)
	}()

	if err != nil {
		collectErrors.Inc(1)
		return nil, err
	}

	src.addSummaryMetricSets(result, summary)
    //
	//podList, err := func() (*kube_api.PodList, error) {
	//	return src.kubeletClient.GetPods(src.node.IP)
	//}()
    //
	//if err != nil {
	//	collectErrors.Inc(1)
	//	return nil, err
	//}
	//src.addCompletedPodMetricSets(result, podList)

	return result, nil
}

const (
	RootFsKey           = "/"
	LogsKey             = "logs"
	NetworkInterfaceKey = "interface_name"
)

// For backwards compatibility, map summary system names into original names.
// TODO: Migrate to the new system names and remove this.
var systemNameMap = map[string]string{
	stats.SystemContainerRuntime: "docker-daemon",
	stats.SystemContainerMisc:    "system",
}

func (src *summaryMetricsSource) addCompletedPodMetricSets(dataBatch *Batch, podList *kube_api.PodList) {
	nodeLabels := map[string]string{
		LabelNodename.Key: src.node.NodeName,
		LabelHostname.Key: src.node.HostName,
		LabelHostID.Key:   src.node.HostID,
	}
	for _, pod := range podList.Items {
		if pod.Status.Phase != kube_api.PodSucceeded && pod.Status.Phase != kube_api.PodFailed {
			continue
		}

		podKey := PodKey(pod.Namespace, pod.Name)
		if dataBatch.Sets[podKey] != nil {
			continue
		}

		podMetrics := &Set{
			Labels:              src.cloneLabels(nodeLabels),
			Values:              map[string]Value{},
			LabeledValues:       []LabeledValue{},
			CollectionStartTime: pod.Status.StartTime.Time,
			ScrapeTime:          dataBatch.Timestamp,
		}

		podMetrics.Labels[LabelMetricSetType.Key] = MetricSetTypePod
		podMetrics.Labels[LabelPodId.Key] = string(pod.UID)
		podMetrics.Labels[LabelPodName.Key] = pod.Name
		podMetrics.Labels[LabelNamespaceName.Key] = pod.Namespace

		dataBatch.Sets[podKey] = podMetrics
		log.Debugf("Added Set for key: %s, status: %s", podKey, pod.Status.Phase)
	}
}

// decodeSummary translates the kubelet statsSummary API into the flattened Set API.
func (src *summaryMetricsSource) addSummaryMetricSets(dataBatch *Batch, summary *stats.Summary) {

	labels := map[string]string{
		LabelNodename.Key: src.node.NodeName,
		LabelHostname.Key: src.node.HostName,
		LabelHostID.Key:   src.node.HostID,
	}

	src.decodeNodeStats(dataBatch.Sets, labels, &summary.Node)
	for _, pod := range summary.Pods {

		src.decodePodStats(dataBatch.Sets, labels, &pod)
	}
	log.Debugf("End summary decode")
}

// Convenience method for labels deep copy.
func (src *summaryMetricsSource) cloneLabels(labels map[string]string) map[string]string {
	clone := make(map[string]string, len(labels))
	for k, v := range labels {
		clone[k] = v
	}
	return clone
}

func (src *summaryMetricsSource) decodeNodeStats(metrics map[ResourceKey]*Set, labels map[string]string, node *stats.NodeStats) {
	log.Tracef("Decoding node stats for node %s...", node.NodeName)
	nodeMetrics := &Set{
		Labels:              src.cloneLabels(labels),
		Values:              map[string]Value{},
		LabeledValues:       []LabeledValue{},
		CollectionStartTime: node.StartTime.Time,
		ScrapeTime:          src.getScrapeTime(node.CPU, node.Memory, node.Network),
	}
	nodeMetrics.Labels[LabelMetricSetType.Key] = MetricSetTypeNode
	nodeMetrics.Labels[LabelNodeRole.Key] = src.node.NodeRole

	src.decodeUptime(nodeMetrics, node.StartTime.Time)
	src.decodeCPUStats(nodeMetrics, node.CPU)
	src.decodeMemoryStats(nodeMetrics, node.Memory)
	src.decodeNetworkStats(nodeMetrics, node.Network)
	src.decodeFsStats(nodeMetrics, RootFsKey, node.Fs)
	src.decodeEphemeralStorageStats(nodeMetrics, node.Fs)
	metrics[NodeKey(node.NodeName)] = nodeMetrics

	for _, container := range node.SystemContainers {
		key := NodeContainerKey(node.NodeName, src.getSystemContainerName(&container))
		containerMetrics := src.decodeContainerStats(labels, &container, true)
		containerMetrics.Labels[LabelMetricSetType.Key] = MetricSetTypeSystemContainer
		metrics[key] = containerMetrics
	}
}

func (src *summaryMetricsSource) decodePodStats(metrics map[ResourceKey]*Set, nodeLabels map[string]string, pod *stats.PodStats) {
	log.Tracef("Decoding pod stats for pod %s/%s (%s)...", pod.PodRef.Namespace, pod.PodRef.Name, pod.PodRef.UID)
	podMetrics := &Set{
		Labels:              src.cloneLabels(nodeLabels),
		Values:              map[string]Value{},
		LabeledValues:       []LabeledValue{},
		CollectionStartTime: pod.StartTime.Time,
		ScrapeTime:          src.getScrapeTime(nil, nil, pod.Network),
	}
	ref := pod.PodRef
	podMetrics.Labels[LabelMetricSetType.Key] = MetricSetTypePod
	podMetrics.Labels[LabelPodId.Key] = ref.UID
	podMetrics.Labels[LabelPodName.Key] = ref.Name
	podMetrics.Labels[LabelNamespaceName.Key] = ref.Namespace

	src.decodeUptime(podMetrics, pod.StartTime.Time)
	src.decodeNetworkStats(podMetrics, pod.Network)
	src.decodeCPUStats(podMetrics, pod.CPU)
	src.decodeMemoryStats(podMetrics, pod.Memory)
	src.decodeEphemeralStorageStats(podMetrics, pod.EphemeralStorage)
	for _, vol := range pod.VolumeStats {
		src.decodeFsStats(podMetrics, VolumeResourcePrefix+vol.Name, &vol.FsStats)
	}
	metrics[PodKey(ref.Namespace, ref.Name)] = podMetrics

	for _, container := range pod.Containers {
		key := PodContainerKey(ref.Namespace, ref.Name, container.Name)
		// This check ensures that we are not replacing metrics of running container with metrics of terminated one if
		// there are two exactly same containers reported by kubelet.
		if _, exist := metrics[key]; exist {
			log.Infof("Metrics reported from two containers with the same key: %v. Create time of "+
				"containers are %v and %v. Metrics from the older container are going to be dropped.", key,
				container.StartTime.Time, metrics[key].CollectionStartTime)
			if container.StartTime.Time.Before(metrics[key].CollectionStartTime) {
				continue
			}
		}
		metrics[key] = src.decodeContainerStats(podMetrics.Labels, &container, false)
	}
}

func (src *summaryMetricsSource) decodeContainerStats(podLabels map[string]string, container *stats.ContainerStats, isSystemContainer bool) *Set {
	log.Tracef("Decoding container stats stats for container %s...", container.Name)
	containerMetrics := &Set{
		Labels:              src.cloneLabels(podLabels),
		Values:              map[string]Value{},
		LabeledValues:       []LabeledValue{},
		CollectionStartTime: container.StartTime.Time,
		ScrapeTime:          src.getScrapeTime(container.CPU, container.Memory, nil),
	}
	containerMetrics.Labels[LabelMetricSetType.Key] = MetricSetTypePodContainer
	if isSystemContainer {
		containerMetrics.Labels[LabelContainerName.Key] = src.getSystemContainerName(container)
	} else {
		containerMetrics.Labels[LabelContainerName.Key] = container.Name
	}

	src.decodeUptime(containerMetrics, container.StartTime.Time)
	src.decodeCPUStats(containerMetrics, container.CPU)
	src.decodeMemoryStats(containerMetrics, container.Memory)
	src.decodeAcceleratorStats(containerMetrics, container.Accelerators)
	src.decodeFsStats(containerMetrics, RootFsKey, container.Rootfs)
	src.decodeFsStats(containerMetrics, LogsKey, container.Logs)
	src.decodeEphemeralStorageStatsForContainer(containerMetrics, container.Rootfs, container.Logs)
	src.decodeUserDefinedMetrics(containerMetrics, container.UserDefinedMetrics)

	return containerMetrics
}

func (src *summaryMetricsSource) decodeUptime(metrics *Set, startTime time.Time) {
	if startTime.IsZero() {
		log.Trace("missing start time!")
		return
	}

	uptime := uint64(time.Since(startTime).Nanoseconds() / time.Millisecond.Nanoseconds())
	src.addIntMetric(metrics, &MetricUptime, &uptime)
}

func (src *summaryMetricsSource) decodeCPUStats(metrics *Set, cpu *stats.CPUStats) {
	if cpu == nil {
		log.Trace("missing cpu usage metric!")
		return
	}
	src.addIntMetric(metrics, &MetricCpuUsage, cpu.UsageCoreNanoSeconds)

	if cpu.UsageNanoCores != nil {
		millicores := *cpu.UsageNanoCores / 1e6
		src.addIntMetric(metrics, &MetricCpuUsageCores, &millicores)
	}
}

func (src *summaryMetricsSource) decodeEphemeralStorageStats(metrics *Set, storage *stats.FsStats) {
	if storage == nil {
		log.Trace("missing storage usage metric!")
		return
	}
	src.addIntMetric(metrics, &MetricEphemeralStorageUsage, storage.UsedBytes)
}

func (src *summaryMetricsSource) decodeEphemeralStorageStatsForContainer(metrics *Set, rootfs *stats.FsStats, logs *stats.FsStats) {
	if rootfs == nil || logs == nil || rootfs.UsedBytes == nil || logs.UsedBytes == nil {
		log.Trace("missing storage usage metric!")
		return
	}
	usage := *rootfs.UsedBytes + *logs.UsedBytes
	src.addIntMetric(metrics, &MetricEphemeralStorageUsage, &usage)
}

func (src *summaryMetricsSource) decodeMemoryStats(metrics *Set, memory *stats.MemoryStats) {
	if memory == nil {
		log.Trace("missing memory metrics!")
		return
	}

	src.addIntMetric(metrics, &MetricMemoryUsage, memory.UsageBytes)
	src.addIntMetric(metrics, &MetricMemoryWorkingSet, memory.WorkingSetBytes)
	src.addIntMetric(metrics, &MetricMemoryRSS, memory.RSSBytes)
	src.addIntMetric(metrics, &MetricMemoryPageFaults, memory.PageFaults)
	src.addIntMetric(metrics, &MetricMemoryMajorPageFaults, memory.MajorPageFaults)
}

func (src *summaryMetricsSource) decodeAcceleratorStats(metrics *Set, accelerators []stats.AcceleratorStats) {
	for _, accelerator := range accelerators {
		acceleratorLabels := map[string]string{
			LabelAcceleratorMake.Key:  accelerator.Make,
			LabelAcceleratorModel.Key: accelerator.Model,
			LabelAcceleratorID.Key:    accelerator.ID,
		}
		src.addLabeledIntMetric(metrics, &MetricAcceleratorMemoryTotal, acceleratorLabels, &accelerator.MemoryTotal)
		src.addLabeledIntMetric(metrics, &MetricAcceleratorMemoryUsed, acceleratorLabels, &accelerator.MemoryUsed)
		src.addLabeledIntMetric(metrics, &MetricAcceleratorDutyCycle, acceleratorLabels, &accelerator.DutyCycle)
	}
}

func (src *summaryMetricsSource) decodeNetworkStats(metrics *Set, network *stats.NetworkStats) {
	if network == nil {
		log.Trace("missing network metrics!")
		return
	}

	for _, netInterface := range network.Interfaces {
		log.Tracef("Processing metrics for network interface %s", netInterface.Name)
		intfLabels := map[string]string{NetworkInterfaceKey: netInterface.Name}
		src.addLabeledIntMetric(metrics, &MetricNetworkRx, intfLabels, netInterface.RxBytes)
		src.addLabeledIntMetric(metrics, &MetricNetworkRxErrors, intfLabels, netInterface.RxErrors)
		src.addLabeledIntMetric(metrics, &MetricNetworkTx, intfLabels, netInterface.TxBytes)
		src.addLabeledIntMetric(metrics, &MetricNetworkTxErrors, intfLabels, netInterface.TxErrors)
	}
	src.addIntMetric(metrics, &MetricNetworkRx, network.RxBytes)
	src.addIntMetric(metrics, &MetricNetworkRxErrors, network.RxErrors)
	src.addIntMetric(metrics, &MetricNetworkTx, network.TxBytes)
	src.addIntMetric(metrics, &MetricNetworkTxErrors, network.TxErrors)
}

func (src *summaryMetricsSource) decodeFsStats(metrics *Set, fsKey string, fs *stats.FsStats) {
	if fs == nil {
		log.Trace("missing fs metrics!")
		return
	}

	fsLabels := map[string]string{LabelResourceID.Key: fsKey}
	src.addLabeledIntMetric(metrics, &MetricFilesystemUsage, fsLabels, fs.UsedBytes)
	src.addLabeledIntMetric(metrics, &MetricFilesystemLimit, fsLabels, fs.CapacityBytes)
	src.addLabeledIntMetric(metrics, &MetricFilesystemAvailable, fsLabels, fs.AvailableBytes)
	src.addLabeledIntMetric(metrics, &MetricFilesystemInodes, fsLabels, fs.Inodes)
	src.addLabeledIntMetric(metrics, &MetricFilesystemInodesFree, fsLabels, fs.InodesFree)
}

func (src *summaryMetricsSource) decodeUserDefinedMetrics(metrics *Set, udm []stats.UserDefinedMetric) {
	for _, metric := range udm {
		metrics.Values[CustomMetricPrefix+metric.Name] = Value{
			ValueType:  ValueFloat,
			FloatValue: metric.Value,
		}
	}
}

func (src *summaryMetricsSource) getScrapeTime(cpu *stats.CPUStats, memory *stats.MemoryStats, network *stats.NetworkStats) time.Time {
	// Assume CPU, memory and network scrape times are the same.
	switch {
	case cpu != nil && !cpu.Time.IsZero():
		return cpu.Time.Time
	case memory != nil && !memory.Time.IsZero():
		return memory.Time.Time
	case network != nil && !network.Time.IsZero():
		return network.Time.Time
	default:
		return time.Time{}
	}
}

// addIntMetric is a convenience method for adding the metric and value to the metric set.
func (src *summaryMetricsSource) addIntMetric(metrics *Set, metric *Metric, value *uint64) {
	if value == nil {
		log.Debugf("skipping metric %s because the value was nil", metric.Name)
		return
	}
	val := Value{
		ValueType: ValueInt64,
		IntValue:  int64(*value),
	}
	metrics.Values[metric.Name] = val
}

// addLabeledIntMetric is a convenience method for adding the labeled metric and value to the metric set.
func (src *summaryMetricsSource) addLabeledIntMetric(metrics *Set, metric *Metric, labels map[string]string, value *uint64) {
	if value == nil {
		log.Debugf("skipping labeled metric %s (%v) because the value was nil", metric.Name, labels)
		return
	}

	val := LabeledValue{
		Name:   metric.Name,
		Labels: labels,
		Value: Value{
			ValueType: ValueInt64,
			IntValue:  int64(*value),
		},
	}
	metrics.LabeledValues = append(metrics.LabeledValues, val)
}

// Translate system container names to the legacy names for backwards compatibility.
func (src *summaryMetricsSource) getSystemContainerName(c *stats.ContainerStats) string {
	if legacyName, ok := systemNameMap[c.Name]; ok {
		return legacyName
	}
	return c.Name
}

type summaryProvider struct {
	metrics.DefaultSourceProvider
	nodeLister       v1listers.NodeLister
	reflector        *cache.Reflector
	kubeletClient    *kubelet.KubeletClient
	hostIDAnnotation string
}

func (sp *summaryProvider) GetMetricsSources() []Source {
	var sources []Source
	nodes, err := sp.nodeLister.List(labels.Everything())
	if err != nil {
		log.Errorf("error while listing nodes: %v", err)
		return sources
	}

	for _, node := range nodes {
		info, err := sp.getNodeInfo(node)
		if err != nil {
			log.Errorf("%v", err)
			continue
		}
		sources = append(sources, NewSummaryMetricsSource(info, sp.kubeletClient))
	}
	return sources
}

func (sp *summaryProvider) Name() string {
	return "kubernetes_summary_provider"
}

func (sp *summaryProvider) getNodeInfo(node *kube_api.Node) (NodeInfo, error) {
	hostname, ip, err := util.GetNodeHostnameAndIP(node)
	if err != nil {
		return NodeInfo{}, err
	}
	hostID := ""
	if sp.hostIDAnnotation != "" {
		hostID = node.Annotations[sp.hostIDAnnotation]
	}
	info := NodeInfo{
		NodeName:       node.Name,
		HostName:       hostname,
		HostID:         hostID,
		IP:             ip,
		KubeletVersion: node.Status.NodeInfo.KubeletVersion,
		NodeRole:       util.GetNodeRole(node),
	}

	log.WithFields(log.Fields{
		"name":      node.Name,
		"hostname":  hostname,
		"hostID":    hostID,
		"ipAddress": ip,
	}).Debug("Node information")

	return info, nil
}

func NewSummaryProvider(cfg configuration.SummarySourceConfig) (SourceProvider, error) {
	hostIDAnnotation := ""

	// create clients
	kubeConfig, kubeletConfig, err := kubelet.GetKubeConfigs(cfg)
	if err != nil {
		return nil, err
	}
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)
	kubeletClient, err := kubelet.NewKubeletClient(kubeletConfig)
	if err != nil {
		return nil, err
	}
	// watch nodes
	nodeLister, reflector, _ := util.GetNodeLister(kubeClient)

	return &summaryProvider{
		nodeLister:       nodeLister,
		reflector:        reflector,
		kubeletClient:    kubeletClient,
		hostIDAnnotation: hostIDAnnotation,
	}, nil
}
