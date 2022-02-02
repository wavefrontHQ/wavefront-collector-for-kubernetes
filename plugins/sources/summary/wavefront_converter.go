// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"strings"

	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// converts MetricSets to MetricPoints.
type pointConverter struct {
	cluster string
	prefix  string
	tags    map[string]string
	filters filter.Filter

	collectedPoints gm.Counter
	filteredPoints  gm.Counter
}

// NewPointConverter creates a new processor that converts summary stats data into the Wavefront point format
func NewPointConverter(cfg configuration.SummarySourceConfig, cluster string) (metrics.DataProcessor, error) {
	cluster = strings.TrimSpace(cluster)
	if cluster == "" {
		cluster = "k8s-cluster"
	}

	pt := map[string]string{"type": "kubernetes.summary_api"}
	return &pointConverter{
		cluster:         cluster,
		prefix:          configuration.GetStringValue(cfg.Prefix, "kubernetes."),
		tags:            cfg.Tags,
		filters:         filter.FromConfig(cfg.Filters),
		collectedPoints: gm.GetOrRegisterCounter(reporting.EncodeKey("source.points.collected", pt), gm.DefaultRegistry),
		filteredPoints:  gm.GetOrRegisterCounter(reporting.EncodeKey("source.points.filtered", pt), gm.DefaultRegistry),
	}, nil
}

func (converter *pointConverter) Name() string {
	return "wavefront_point_converter"
}

func (converter *pointConverter) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	if len(batch.MetricSets) == 0 {
		return batch, nil
	}

	metricSets := batch.MetricSets
	nodeName := util.GetNodeName()
	ts := batch.Timestamp

	log.WithField("total", len(metricSets)).Debug("Processing metric sets")

	for _, key := range sortedMetricSetKeys(metricSets) {
		ms := metricSets[key]

		// Populate tag map
		tags := make(map[string]string)

		// Add pod labels as tags
		converter.addLabelTags(ms, tags)
		hostname := tags["hostname"]
		metricType := tags["type"]
		if strings.Contains(tags["container_name"], sysSubContainerName) {
			//don't send system subcontainers
			continue
		}
		for _, metricName := range sortedMetricValueKeys(ms.MetricValues) {
			metricValue := ms.MetricValues[metricName]
			var value float64
			if metrics.ValueInt64 == metricValue.ValueType {
				value = float64(metricValue.IntValue)
			} else if metrics.ValueFloat == metricValue.ValueType {
				value = metricValue.FloatValue
			} else {
				continue
			}

			ts := ts.Unix()
			source := nodeName
			if source == "" {
				if metricType == "cluster" {
					source = converter.cluster
				} else if metricType == "ns" {
					source = tags["namespace_name"] + "-ns"
				} else {
					source = hostname
				}
			}

			// convert to a point and add it to the data batch
			point := converter.metricPoint(converter.cleanMetricName(metricType, metricName), value, ts, source, tags)
			batch.MetricPoints = converter.filterAppend(batch.MetricPoints, point)
			converter.collectedPoints.Inc(1)
		}
		for _, metric := range ms.LabeledMetrics {
			var value float64
			if metrics.ValueInt64 == metric.ValueType {
				value = float64(metric.IntValue)
			} else if metrics.ValueFloat == metric.ValueType {
				value = metric.FloatValue
			} else {
				continue
			}

			ts := ts.Unix()
			source := nodeName
			if source == "" {
				source = hostname
			}
			labels := metric.Labels
			if labels == nil {
				labels = make(map[string]string, len(tags))
			}
			for k, v := range tags {
				labels[k] = v
			}

			// convert to a point and add it to the data batch
			point := converter.metricPoint(converter.cleanMetricName(metricType, metric.Name), value, ts, source, labels)
			batch.MetricPoints = converter.filterAppend(batch.MetricPoints, point)
			converter.collectedPoints.Inc(1)
		}
	}
	return batch, nil
}

func (converter *pointConverter) filterAppend(slice []*metrics.MetricPoint, point *metrics.MetricPoint) []*metrics.MetricPoint {
	if converter.filters == nil || converter.filters.MatchMetric(point.Metric, point.GetTags()) {
		return append(slice, point)
	}
	converter.filteredPoints.Inc(1)
	if log.IsLevelEnabled(log.TraceLevel) {
		log.WithField("name", point.Metric).Trace("Dropping metric")
	}
	return slice
}

func (converter *pointConverter) addLabelTags(ms *metrics.MetricSet, tags map[string]string) {
	for _, labelName := range sortedLabelKeys(ms.Labels) {
		labelValue := ms.Labels[labelName]
		if labelName == "labels" {
			for _, label := range strings.Split(labelValue, ",") {
				//labels = app:webproxy,version:latest
				tagParts := strings.SplitN(label, ":", 2)
				if len(tagParts) == 2 {
					tags["label."+tagParts[0]] = tagParts[1]
				}
			}
		} else {
			tags[labelName] = labelValue
		}
	}
}

func (converter *pointConverter) metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	point := &metrics.MetricPoint{
		Metric:    name,
		Value:     value,
		Timestamp: ts,
		Source:    source,
	}
	point.SetTags(tags)
	return point
}

func (converter *pointConverter) cleanMetricName(metricType string, metricName string) string {
	return converter.prefix + metricType + "." + strings.Replace(metricName, "/", ".", -1)
}
