package stats

import (
	"fmt"
	"strings"
	"time"

	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

var (
	source          string
	filters         []string
	pointsCollected metrics.Counter
)

func init() {
	source = util.GetNodeName()
	if source == "" {
		source = "wavefront-kubernetes-collector"
	}

	// filter out if zero valued
	filters = []string{
		"filtered.count",
		"errors.count",
		"targets.registered",
		"collect.errors",
		"points.filtered",
		"points.collected",
	}

	// internal stats pps
	name := reporting.EncodeKey("source.points.collected", map[string]string{"type": "internal"})
	pointsCollected = metrics.GetOrRegisterCounter(name, metrics.DefaultRegistry)
}

func internalStats() (*DataBatch, error) {
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
			points = filterAppend(points, point(name, float64(metric.Count()), now.Unix(), source))
		case metrics.Gauge:
			points = filterAppend(points, point(name, float64(metric.Value()), now.Unix(), source))
		case metrics.GaugeFloat64:
			points = append(points, point(name, metric.Value(), now.Unix(), source))
		case metrics.Timer:
			timer := metric.Snapshot()
			points = append(points, addHisto(name, timer.Min(), timer.Max(), timer.Mean(),
				timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
			points = append(points, addRate(name, timer.Count(), timer.Rate1(), timer.RateMean(), now.Unix())...)
		case metrics.Histogram:
			histo := metric.Snapshot()
			points = append(points, addHisto(name, histo.Min(), histo.Max(), histo.Mean(),
				histo.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())...)
		case metrics.Meter:
			meter := metric.Snapshot()
			points = append(points, addRate(name, meter.Count(), meter.Rate1(), meter.RateMean(), now.Unix())...)
		}
	})
	pointsCollected.Inc(int64(len(points)))
	result.MetricPoints = points
	return result, nil
}

func addHisto(name string, min, max int64, mean float64, percentiles []float64, now int64) []*MetricPoint {
	// convert from nanoseconds to milliseconds
	var points []*MetricPoint
	points = append(points, point(combine(name, "duration.min"), float64(min)/1e6, now, source))
	points = append(points, point(combine(name, "duration.max"), float64(max)/1e6, now, source))
	points = append(points, point(combine(name, "duration.mean"), mean/1e6, now, source))
	points = append(points, point(combine(name, "duration.median"), percentiles[0]/1e6, now, source))
	points = append(points, point(combine(name, "duration.p75"), percentiles[1]/1e6, now, source))
	points = append(points, point(combine(name, "duration.p95"), percentiles[2]/1e6, now, source))
	points = append(points, point(combine(name, "duration.p99"), percentiles[3]/1e6, now, source))
	points = append(points, point(combine(name, "duration.p999"), percentiles[4]/1e6, now, source))
	return points
}

func addRate(name string, count int64, m1, mean float64, now int64) []*MetricPoint {
	var points []*MetricPoint
	points = append(points, point(combine(name, "rate.count"), float64(count), now, source))
	points = append(points, point(combine(name, "rate.m1"), m1, now, source))
	points = append(points, point(combine(name, "rate.mean"), mean, now, source))
	return points
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}

func point(name string, value float64, ts int64, source string) *MetricPoint {
	name, tags := reporting.DecodeKey(name)
	if filterName(name) && value == 0.0 {
		// don't emit internal counts with zero values
		return nil
	}

	return &MetricPoint{
		Metric:    statsPrefix + "collector." + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

func filterAppend(slice []*MetricPoint, point *MetricPoint) []*MetricPoint {
	if point == nil {
		return slice
	}
	return append(slice, point)
}

func filterName(name string) bool {
	for _, filter := range filters {
		if strings.HasSuffix(name, filter) {
			return true
		}
	}
	return false
}
