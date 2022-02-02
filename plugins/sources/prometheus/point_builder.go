// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"math"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

type pointBuilder struct {
	isValidMetric    func(name string, tags map[string]string) bool
	source           string
	prefix           string
	omitBucketSuffix bool
	tags             map[string]string
	interner         util.StringInterner
}

func NewPointBuilder(src *prometheusMetricsSource) *pointBuilder {
	return &pointBuilder{
		source:           src.source,
		prefix:           src.prefix,
		omitBucketSuffix: src.omitBucketSuffix,
		tags:             src.tags,
		isValidMetric:    src.isValidMetric,
		interner:         util.NewStringInterner(),
	}

}

// build converts a map of prometheus metric families by metric name to a collection of wavefront points
// build actually never returns an error
func (builder *pointBuilder) build(metricFamilies map[string]*dto.MetricFamily) ([]*metrics.MetricPoint, error) {
	now := time.Now().Unix()
	var result []*metrics.MetricPoint

	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			var points []*metrics.MetricPoint
			// Prometheus metric family -> wavefront metric points
			if mf.GetType() == dto.MetricType_SUMMARY {
				points = builder.buildSummaryPoints(metricName, m, now, builder.buildTags(m))
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				points = builder.buildHistogramPoints(metricName, m, now, builder.buildTags(m))
			} else {
				points = builder.buildPoints(metricName, m, now)
			}

			if len(points) > 0 {
				result = append(result, points...)
			}
		}
	}
	return result, nil
}

func (builder *pointBuilder) metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	point := &metrics.MetricPoint{
		Metric:    builder.prefix + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
	}
	point.SetLabelPairs(builder.deduplicate(tags)) //store tags as LabelPairs for memory optimization
	return point
}

func (builder *pointBuilder) filterAppend(slice []*metrics.MetricPoint, point *metrics.MetricPoint) []*metrics.MetricPoint {
	if builder.isValidMetric(point.Metric, point.GetTags()) {
		return append(slice, point)
	}
	return slice
}

// Get name and value from metric
func (builder *pointBuilder) buildPoints(name string, m *dto.Metric, now int64) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			point := builder.metricPoint(name+".gauge", m.GetGauge().GetValue(), now, builder.source, builder.buildTags(m))
			result = builder.filterAppend(result, point)
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			point := builder.metricPoint(name+".counter", m.GetCounter().GetValue(), now, builder.source, builder.buildTags(m))
			result = builder.filterAppend(result, point)
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			point := builder.metricPoint(name+".value", m.GetUntyped().GetValue(), now, builder.source, builder.buildTags(m))
			result = builder.filterAppend(result, point)
		}
	}
	return result
}

// Get Quantiles from summary metric
func (builder *pointBuilder) buildSummaryPoints(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			newTags := copyOf(tags)
			newTags["quantile"] = fmt.Sprintf("%v", q.GetQuantile())
			point := builder.metricPoint(name, q.GetValue(), now, builder.source, newTags)
			result = builder.filterAppend(result, point)
		}
	}
	point := builder.metricPoint(name+".count", float64(m.GetSummary().GetSampleCount()), now, builder.source, tags)
	result = builder.filterAppend(result, point)
	point = builder.metricPoint(name+".sum", m.GetSummary().GetSampleSum(), now, builder.source, tags)
	result = builder.filterAppend(result, point)

	return result
}

// Get Buckets from histogram metric
func (builder *pointBuilder) buildHistogramPoints(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	histName := builder.histogramName(name)
	for _, b := range m.GetHistogram().Bucket {
		newTags := copyOf(tags)
		newTags["le"] = fmt.Sprintf("%v", b.GetUpperBound())
		point := builder.metricPoint(histName, float64(b.GetCumulativeCount()), now, builder.source, newTags)
		result = builder.filterAppend(result, point)
	}
	point := builder.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, builder.source, tags)
	result = builder.filterAppend(result, point)
	point = builder.metricPoint(name+".sum", m.GetHistogram().GetSampleSum(), now, builder.source, tags)
	result = builder.filterAppend(result, point)
	return result
}

// Get labels from metric
func (builder *pointBuilder) buildTags(m *dto.Metric) map[string]string {
	tags := make(map[string]string, len(builder.tags)+len(m.Label))
	for k, v := range builder.tags {
		if len(v) > 0 {
			tags[k] = v
		}
	}
	if len(m.Label) >= 0 {
		for _, label := range m.Label {
			if len(label.GetName()) > 0 && len(label.GetValue()) > 0 {
				tags[label.GetName()] = label.GetValue()
			}
		}
	}
	return tags
}

func (builder *pointBuilder) histogramName(name string) string {
	if builder.omitBucketSuffix {
		return name
	}
	return name + ".bucket"
}

func (builder *pointBuilder) deduplicate(tags map[string]string) []metrics.LabelPair {
	result := make([]metrics.LabelPair, 0)
	for k, v := range tags {
		result = append(result, metrics.LabelPair{
			Name:  builder.interner.Intern(k),
			Value: builder.interner.Intern(v),
		})
	}
	return result
}

func copyOf(tags map[string]string) map[string]string {
	newTags := make(map[string]string, len(tags)+1)
	for k, v := range tags {
		newTags[k] = v
	}
	return newTags
}
