// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stats

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"

	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

type internalMetricsSource struct {
	metrics.DefaultSourceProvider
	prefix  string
	tags    map[string]string
	filters filter.Filter

	source      string
	zeroFilters []string
	pps         gometrics.Counter
	fps         gometrics.Counter
}

func newInternalMetricsSource(prefix string, tags map[string]string, filters filter.Filter) (metrics.Source, error) {
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
		prefix:  prefix,
		tags:    tags,
		filters: filters,

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

func (src *internalMetricsSource) AutoDiscovered() bool {
	return false
}

func (src *internalMetricsSource) Name() string {
	return "internal_stats_source"
}

func (src *internalMetricsSource) Cleanup() {}

func (src *internalMetricsSource) Scrape() (*metrics.Batch, error) {
	return src.internalStats()
}

func (src *internalMetricsSource) internalStats() (*metrics.Batch, error) {
	now := time.Now()
	result := &metrics.Batch{
		Timestamp: now,
	}
	var points []wf.Metric

	src.tags["leading"] = strconv.FormatBool(leadership.Leading())
	src.tags["installation_method"] = util.GetInstallationMethod()
	util.AddK8sTags(src.tags)

	// update GC and memory stats before populating the map
	gometrics.CaptureRuntimeMemStatsOnce(gometrics.DefaultRegistry)

	gometrics.DefaultRegistry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case gometrics.Counter:
			points = wf.FilterAppend(src.filters, src.fps, points, src.point(name, float64(metric.Count()), now.Unix()))
		case gometrics.Gauge:
			points = wf.FilterAppend(src.filters, src.fps, points, src.point(name, float64(metric.Value()), now.Unix()))
		case gometrics.GaugeFloat64:
			points = wf.FilterAppend(src.filters, src.fps, points, src.point(name, metric.Value(), now.Unix()))
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
	result.Metrics = points
	return result, nil
}

func (src *internalMetricsSource) addHisto(name string, min, max int64, mean float64, percentiles []float64, now int64) []wf.Metric {
	// convert from nanoseconds to milliseconds
	var points []wf.Metric
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.min"), float64(min)/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.max"), float64(max)/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.mean"), mean/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.median"), percentiles[0]/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.p75"), percentiles[1]/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.p95"), percentiles[2]/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.p99"), percentiles[3]/1e6, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "duration.p999"), percentiles[4]/1e6, now))
	return points
}

func (src *internalMetricsSource) addRate(name string, count int64, m1, mean float64, now int64) []wf.Metric {
	var points []wf.Metric
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "rate.count"), float64(count), now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "rate.m1"), m1, now))
	points = wf.FilterAppend(src.filters, src.fps, points, src.point(combine(name, "rate.mean"), mean, now))
	return points
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}

func (src *internalMetricsSource) point(name string, value float64, ts int64) wf.Metric {
	name, tags := reporting.DecodeKey(name)
	if value == 0.0 && src.filterName(name) {
		// don't emit internal counts with zero values
		return nil
	}
	return wf.NewPoint(
		src.prefix+"collector."+strings.Replace(name, "_", ".", -1),
		value,
		ts,
		src.source,
		src.buildTags(tags),
	)
}

func (src *internalMetricsSource) buildTags(tags map[string]string) map[string]string {
	for k, v := range src.tags {
		if len(v) > 0 {
			if _, exists := tags[k]; !exists {
				tags[k] = v
			}
		}
	}
	return tags
}

func (src *internalMetricsSource) filterName(name string) bool {
	for _, suffix := range src.zeroFilters {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
