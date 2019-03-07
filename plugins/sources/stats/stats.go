package stats

import (
	"fmt"
	"strings"
	"time"

	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/rcrowley/go-metrics"
)

const source = "wavefront-kubernetes-collector"

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
			points = append(points, point(name, float64(metric.Count()), now.Unix(), source, nil))
		case metrics.Gauge:
			points = append(points, point(name, float64(metric.Value()), now.Unix(), source, nil))
		case metrics.GaugeFloat64:
			points = append(points, point(name, metric.Value(), now.Unix(), source, nil))
		case metrics.Timer:
			timer := metric.Snapshot()
			addHisto(name, timer.Min(), timer.Max(), timer.Mean(),
				timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())
			addRate(name, timer.Count(), timer.Rate1(), timer.RateMean(), now.Unix())
		case metrics.Histogram:
			histo := metric.Snapshot()
			addHisto(name, histo.Min(), histo.Max(), histo.Mean(),
				histo.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999}), now.Unix())
		case metrics.Meter:
			meter := metric.Snapshot()
			addRate(name, meter.Count(), meter.Rate1(), meter.RateMean(), now.Unix())
		}
	})
	result.MetricPoints = points
	return result, nil
}

func addHisto(name string, min, max int64, mean float64, percentiles []float64, now int64) []*MetricPoint {
	// convert from nanoseconds to milliseconds
	var points []*MetricPoint
	points = append(points, point(combine(name, "duration.min"), float64(min)/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.max"), float64(max)/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.mean"), mean/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.median"), percentiles[0]/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.p75"), percentiles[1]/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.p95"), percentiles[2]/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.p99"), percentiles[3]/1e6, now, source, nil))
	points = append(points, point(combine(name, "duration.p999"), percentiles[4]/1e6, now, source, nil))
	return points
}

func addRate(name string, count int64, m1, mean float64, now int64) []*MetricPoint {
	var points []*MetricPoint
	points = append(points, point(combine(name, "rate.count"), float64(count), now, source, nil))
	points = append(points, point(combine(name, "rate.m1"), m1, now, source, nil))
	points = append(points, point(combine(name, "rate.mean"), mean, now, source, nil))
	return points
}

func combine(prefix, name string) string {
	return fmt.Sprintf("%s.%s", prefix, name)
}

func point(name string, value float64, ts int64, source string, tags map[string]string) *MetricPoint {
	return &MetricPoint{
		Metric:    statsPrefix + "collector." + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}
