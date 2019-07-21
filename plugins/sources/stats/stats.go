package stats

import (
	"fmt"
	"strings"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

type internalMetricsSource struct {
	DefaultMetricsSourceProvider
	prefix  string
	tags    map[string]string
	filters filter.Filter

	source      string
	zeroFilters []string
	pps         metrics.Counter
	fps         metrics.Counter
}

func newInternalMetricsSource(prefix string, tags map[string]string, filters filter.Filter) (MetricsSource, error) {
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
	return &internalMetricsSource{
		prefix:  prefix,
		tags:    tags,
		filters: filters,

		zeroFilters: zeroFilters,
		source:      getDefault(util.GetNodeName(), "wavefront-kubernetes-collector"),
		pps:         metrics.GetOrRegisterCounter(ppsKey, metrics.DefaultRegistry),
		fps:         metrics.GetOrRegisterCounter(fpsKey, metrics.DefaultRegistry),
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

func (src *internalMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	return src.internalStats()
}

func (src *internalMetricsSource) internalStats() (*DataBatch, error) {
	now := time.Now()
	result := &DataBatch{
		Timestamp: now,
	}
	var points []*MetricPoint

	// update GC and memory stats before populating the map
	metrics.CaptureRuntimeMemStatsOnce(metrics.DefaultRegistry)

	metrics.DefaultRegistry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			points = src.filterAppend(points, src.point(name, float64(metric.Count()), now.Unix()))
		case metrics.Gauge:
			points = src.filterAppend(points, src.point(name, float64(metric.Value()), now.Unix()))
		case metrics.GaugeFloat64:
			points = src.filterAppend(points, src.point(name, metric.Value(), now.Unix()))
		case metrics.Timer:
			timer := metric.Snapshot()
			points = append(points, src.addHisto(name, timer.Min(), timer.Max(), timer.Mean(),
				timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
			points = append(points, src.addRate(name, timer.Count(), timer.Rate1(), timer.RateMean(), now.Unix())...)
		case metrics.Histogram:
			histo := metric.Snapshot()
			points = append(points, src.addHisto(name, histo.Min(), histo.Max(), histo.Mean(),
				histo.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
		case metrics.Meter:
			meter := metric.Snapshot()
			points = append(points, src.addRate(name, meter.Count(), meter.Rate1(), meter.RateMean(), now.Unix())...)
		}
	})
	src.pps.Inc(int64(len(points)))
	result.MetricPoints = points
	return result, nil
}

func (src *internalMetricsSource) addHisto(name string, min, max int64, mean float64, percentiles []float64, now int64) []*MetricPoint {
	// convert from nanoseconds to milliseconds
	var points []*MetricPoint
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

func (src *internalMetricsSource) addRate(name string, count int64, m1, mean float64, now int64) []*MetricPoint {
	var points []*MetricPoint
	points = src.filterAppend(points, src.point(combine(name, "rate.count"), float64(count), now))
	points = src.filterAppend(points, src.point(combine(name, "rate.m1"), m1, now))
	points = src.filterAppend(points, src.point(combine(name, "rate.mean"), mean, now))
	return points
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}

func (src *internalMetricsSource) point(name string, value float64, ts int64) *MetricPoint {
	name, tags := reporting.DecodeKey(name)
	if value == 0.0 && src.filterName(name) {
		// don't emit internal counts with zero values
		return nil
	}

	return &MetricPoint{
		Metric:    src.prefix + "collector." + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    src.source,
		Tags:      src.buildTags(tags),
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

func (src *internalMetricsSource) filterAppend(slice []*MetricPoint, point *MetricPoint) []*MetricPoint {
	if point == nil {
		return slice
	}
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
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
