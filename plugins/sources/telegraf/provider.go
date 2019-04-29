package telegraf

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/influxdata/telegraf"
	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	wf "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

// NewProvider creates a Telegraf source
func NewProvider(uri *url.URL) (wf.MetricsSourceProvider, error) {
	glog.Infof("[telegraf.NewProvider] - inputs: %v -------------", telegrafInputs.Inputs)

	var sources []wf.MetricsSource
	for plugin, creator := range telegrafInputs.Inputs {
		sources = append(sources, newTelegrafPluginSource(plugin+" plugin", creator()))
	}

	return &telegrafProvider{sources: sources}, nil
}

type telegrafProvider struct {
	sources []wf.MetricsSource
}

func (p telegrafProvider) GetMetricsSources() []wf.MetricsSource {
	return p.sources
}

func (p telegrafProvider) Name() string {
	return "Telegraf Source"
}

type telegrafPluginSource struct {
	name   string
	source string
	plugin telegraf.Input
	points []*wf.MetricPoint
	mux    sync.Mutex
}

func newTelegrafPluginSource(name string, plugin telegraf.Input) *telegrafPluginSource {
	hostname, err := os.Hostname()
	if err != nil {
		glog.Fatal(err)
	}

	glog.Infof("hostname: '%s'", hostname)

	tsp := &telegrafPluginSource{name: name, plugin: plugin, source: hostname}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := plugin.Gather(tsp)
				if err != nil {
					glog.Errorf("Error on Gather - plugin '%s' - error: %v", name, err)
				}
			}
		}
	}()

	return tsp
}

func (t *telegrafPluginSource) Name() string {
	return t.name + " plugin Source"
}

func (t *telegrafPluginSource) ScrapeMetrics(start, end time.Time) (*wf.DataBatch, error) {
	result := &wf.DataBatch{
		Timestamp: time.Now(),
	}

	t.mux.Lock()
	result.MetricPoints = t.points
	t.points = nil
	t.mux.Unlock()

	return result, nil
}

func (t *telegrafPluginSource) preparePoints(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	var points []*wf.MetricPoint
	var ts time.Time
	if len(timestamp) > 0 {
		ts = timestamp[0]
	} else {
		ts = time.Now()
	}

	for metric, v := range fields {
		var value float64
		switch v.(type) {
		case int:
			value = float64(v.(int))
		case uint64:
			value = float64(v.(uint64))
		case int64:
			value = float64(v.(int64))
		case float64:
			value = v.(float64)
		default:
			glog.Errorf("Error, unsupported type '%v'", reflect.TypeOf(v))
			continue
		}

		point := &wf.MetricPoint{
			Metric:    measurement + "." + strings.Replace(metric, "_", ".", -1),
			Value:     value,
			Timestamp: ts.UnixNano() / 1000,
			Source:    t.source,
			Tags:      tags,
		}
		points = append(points, point)
	}
	t.mux.Lock()
	t.points = append(t.points, points...)
	t.mux.Unlock()
}

// AddFields adds a metric to the accumulator with the given measurement
// name, fields, and tags (and timestamp). If a timestamp is not provided,
// then the accumulator sets it to "now".
func (t *telegrafPluginSource) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddGauge is the same as AddFields, but will add the metric as a "Gauge" type
func (t *telegrafPluginSource) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddCounter is the same as AddFields, but will add the metric as a "Counter" type
func (t *telegrafPluginSource) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddSummary is the same as AddFields, but will add the metric as a "Summary" type
func (t *telegrafPluginSource) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddHistogram is the same as AddFields, but will add the metric as a "Histogram" type
func (t *telegrafPluginSource) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddMetric adds an metric to the accumulator.
func (t *telegrafPluginSource) AddMetric(telegraf.Metric) {
	glog.Fatal("not supported")
}

// SetPrecision sets the timestamp rounding precision.  All metrics addeds
// added to the accumulator will have their timestamp rounded to the
// nearest multiple of precision.
func (t *telegrafPluginSource) SetPrecision(precision time.Duration) {
	glog.Fatal("not supported")
}

// Report an error.
func (t *telegrafPluginSource) AddError(err error) {
	glog.Fatal("not supported")
}

// Upgrade to a TrackingAccumulator with space for maxTracked
// metrics/batches.
func (t *telegrafPluginSource) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	glog.Fatal("not supported")
	return nil
}
