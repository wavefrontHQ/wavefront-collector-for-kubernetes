// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stats

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"

	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

type internalMetricsSource struct {
	metrics.DefaultMetricsSourceProvider
	prefix      string
	tags        map[string]string
	filters     filter.Filter
	tagsEncoder util.TagsEncoder

	source      string
	zeroFilters []string
	pps         gometrics.Counter
	fps         gometrics.Counter
}

func newInternalMetricsSource(prefix string, tags map[string]string, filters filter.Filter) (metrics.MetricsSource, error) {
	ppsKey := reporting.EncodeKey("source.points.collected", map[string]string{"type": "internal"})
	fpsKey := reporting.EncodeKey("source.points.filtered", map[string]string{"type": "internal"})

	zeroFilters := []string{
		"filtered.count",
		"errors.count",
		"targets.registered",
		"collect.errors",
		"points.filtered",
		"points.collected",
	}
	if len(tags) == 0 {
		tags = make(map[string]string, 1)
	}
	return &internalMetricsSource{
		prefix:      prefix,
		tags:        tags,
		filters:     filters,
		tagsEncoder: util.NewTagsEncoder(),
		zeroFilters: zeroFilters,
		source:      getDefault(util.GetNodeName(), "wavefront-collector-for-kubernetes"),
		pps:         gometrics.GetOrRegisterCounter(ppsKey, gometrics.DefaultRegistry),
		fps:         gometrics.GetOrRegisterCounter(fpsKey, gometrics.DefaultRegistry),
	}, nil
}

func getDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func (src *internalMetricsSource) Name() string {
	return "internal_stats_source"
}

func (src *internalMetricsSource) ScrapeMetrics() (*metrics.DataBatch, error) {
	return src.internalStats()
}

func (src *internalMetricsSource) internalStats() (*metrics.DataBatch, error) {
	now := time.Now()
	result := &metrics.DataBatch{
		Timestamp: now,
	}
	var points []*metrics.MetricPointWithStrTags

	src.tags["leading"] = strconv.FormatBool(leadership.Leading())

	// update GC and memory stats before populating the map
	gometrics.CaptureRuntimeMemStatsOnce(gometrics.DefaultRegistry)

	gometrics.DefaultRegistry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case gometrics.Counter:
			points = src.filterAppend(points, src.point(name, float64(metric.Count()), now.Unix()))
		case gometrics.Gauge:
			points = src.filterAppend(points, src.point(name, float64(metric.Value()), now.Unix()))
		case gometrics.GaugeFloat64:
			points = src.filterAppend(points, src.point(name, metric.Value(), now.Unix()))
		case gometrics.Timer:
			timer := metric.Snapshot()
			points = append(points, src.addHisto(name, timer.Min(), timer.Max(), timer.Mean(),
				timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
			points = append(points, src.addRate(name, timer.Count(), timer.Rate1(), timer.RateMean(), now.Unix())...)
		case gometrics.Histogram:
			histo := metric.Snapshot()
			points = append(points, src.addHisto(name, histo.Min(), histo.Max(), histo.Mean(),
				histo.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
		case gometrics.Meter:
			meter := metric.Snapshot()
			points = append(points, src.addRate(name, meter.Count(), meter.Rate1(), meter.RateMean(), now.Unix())...)
		}
	})
	src.pps.Inc(int64(len(points)))
	result.MetricPoints = points
	return result, nil
}

func (src *internalMetricsSource) addHisto(name string, min, max int64, mean float64, percentiles []float64, now int64) []*metrics.MetricPointWithStrTags {
	// convert from nanoseconds to milliseconds
	var points []*metrics.MetricPointWithStrTags
	points = src.filterAppend(points, src.point(combine(name, "duration.min"), float64(min)/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.max"), float64(max)/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.mean"), mean/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.median"), percentiles[0]/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.p75"), percentiles[1]/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.p95"), percentiles[2]/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.p99"), percentiles[3]/1e6, now))
	points = src.filterAppend(points, src.point(combine(name, "duration.p999"), percentiles[4]/1e6, now))
	return points
}

func (src *internalMetricsSource) addRate(name string, count int64, m1, mean float64, now int64) []*metrics.MetricPointWithStrTags {
	var points []*metrics.MetricPointWithStrTags
	points = src.filterAppend(points, src.point(combine(name, "rate.count"), float64(count), now))
	points = src.filterAppend(points, src.point(combine(name, "rate.m1"), m1, now))
	points = src.filterAppend(points, src.point(combine(name, "rate.mean"), mean, now))
	return points
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}

func (src *internalMetricsSource) point(name string, value float64, ts int64) *metrics.MetricPointWithTags {
	name, tags := reporting.DecodeKey(name)
	if value == 0.0 && src.filterName(name) {
		// don't emit internal counts with zero values
		return nil
	}

	return &metrics.MetricPointWithTags{
		MetricPoint: metrics.MetricPoint{
			Metric:    src.prefix + "collector." + strings.Replace(name, "_", ".", -1),
			Value:     value,
			Timestamp: ts,
			Source:    src.source},
		Tags: src.buildTags(tags),
	}
}

func (src *internalMetricsSource) buildTags(tags map[string]string) map[string]string {
	if len(src.tags) == 0 {
		return tags
	}
	if len(tags) == 0 {
		return src.tags
	}
	for k, v := range src.tags {
		if len(v) > 0 {
			if _, exists := tags[k]; !exists {
				tags[k] = v
			}
		}
	}
	return tags
}

func (src *internalMetricsSource) filterAppend(slice []*metrics.MetricPointWithStrTags, point *metrics.MetricPointWithTags) []*metrics.MetricPointWithStrTags {
	if point == nil {
		return slice
	}
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		newPoint := &metrics.MetricPointWithStrTags{
			MetricPoint: point.MetricPoint,
			StrTags:     src.tagsEncoder.Encode(point.Tags),
		}
		return append(slice, newPoint)
	}
	src.fps.Inc(1)
	return slice
}

func (src *internalMetricsSource) filterName(name string) bool {
	for _, suffix := range src.zeroFilters {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
