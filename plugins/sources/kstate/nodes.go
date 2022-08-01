// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	v1 "k8s.io/api/core/v1"
)

func pointsForNode(item interface{}, transforms configuration.Transforms) []wf.Metric {
	node, ok := item.(*v1.Node)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}
	now := time.Now().Unix()
	points := buildNodeConditions(node, transforms, now)
	points = append(points, buildNodeTaints(node, transforms, now)...)
	points = append(points, buildNodeInfo(node, transforms, now))
	return points
}

func buildNodeConditions(node *v1.Node, transforms configuration.Transforms, ts int64) []wf.Metric {
	points := make([]wf.Metric, len(node.Status.Conditions))
	for i, condition := range node.Status.Conditions {
		tags := buildTags("nodename", node.Name, "", transforms.Tags)
		copyLabels(node.GetLabels(), tags)
		tags["status"] = string(condition.Status)
		tags["condition"] = string(condition.Type)
		tags[metrics.LabelNodeRole.Key] = util.GetNodeRole(node)

		// add status and condition (condition.status and condition.type)
		points[i] = metricPoint(transforms.Prefix, "node.status.condition",
			nodeConditionFloat64(condition.Status), ts, transforms.Source, tags)
	}
	return points
}

func buildNodeTaints(node *v1.Node, transforms configuration.Transforms, ts int64) []wf.Metric {
	points := make([]wf.Metric, len(node.Spec.Taints))
	for i, taint := range node.Spec.Taints {
		tags := buildTags("nodename", node.Name, "", transforms.Tags)
		copyLabels(node.GetLabels(), tags)
		tags["key"] = taint.Key
		tags["value"] = taint.Value
		tags["effect"] = string(taint.Effect)
		points[i] = metricPoint(transforms.Prefix, "node.spec.taint", 1.0, ts, transforms.Source, tags)
	}
	return points
}

func buildNodeInfo(node *v1.Node, transforms configuration.Transforms, ts int64) wf.Metric {
	tags := buildTags("nodename", node.Name, "", transforms.Tags)
	copyLabels(node.GetLabels(), tags)
	tags["kernel_version"] = node.Status.NodeInfo.KernelVersion
	tags["os_image"] = node.Status.NodeInfo.OSImage
	tags["container_runtime_version"] = node.Status.NodeInfo.ContainerRuntimeVersion
	tags["kubelet_version"] = node.Status.NodeInfo.KubeletVersion
	tags["kubeproxy_version"] = node.Status.NodeInfo.KubeProxyVersion
	tags["provider_id"] = node.Spec.ProviderID
	tags["pod_cidr"] = node.Spec.PodCIDR
	tags["node_role"] = util.GetNodeRole(node)

	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			tags["internal_ip"] = address.Address
		}
	}
	return metricPoint(transforms.Prefix, "node.info", 1.0, ts, transforms.Source, tags)
}

func nodeConditionFloat64(status v1.ConditionStatus) float64 {
	switch status {
	case v1.ConditionTrue:
		return 1.0
	case v1.ConditionFalse:
		return 0.0
	default:
		return -1.0
	}
}
