package telegraf

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/influxdata/telegraf"
	telegrafPlugins "github.com/influxdata/telegraf/plugins/inputs"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

// NewProvider creates a Telegraf source
func NewProvider(uri *url.URL) (metrics.MetricsSourceProvider, error) {
	for _, pair := range os.Environ() {
		glog.Infof("env: %v", pair)
	}

	glog.Errorf("[NewProvider] uri: '%v'", uri)
	vals := uri.Query()
	glog.Infof("[NewProvider] vals: '%v'", vals)

	prefix := ""
	if len(vals["prefix"]) > 0 {
		prefix = vals["prefix"][0]
		prefix = strings.Trim(prefix, ".")
	}

	var plugins []string
	for _, pluginList := range vals["plugins"] {
		plugins = append(plugins, strings.Split(pluginList, ",")...)
	}
	if len(plugins) == 0 {
		plugins = []string{"mem", "net", "netstat", "linux_sysctl_fs", "swap", "cpu", "disk", "diskio", "system", "kernel", "processes"}
	}

	var sources []metrics.MetricsSource
	for _, name := range plugins {
		creator := telegrafPlugins.Inputs[strings.Trim(name, " ")]
		if creator != nil {
			sources = append(sources, newTelegrafPluginSource(name+" plugin", creator(), prefix))
		} else {
			glog.Errorf("[NewProvider] Error, plugin '%v' not Found", name)
			var availablePlugins []string
			for name := range telegrafPlugins.Inputs {
				availablePlugins = append(availablePlugins, name)
			}
			glog.Infof("[NewProvider] Available plugins: '%v'", availablePlugins)
		}
	}

	return &telegrafProvider{sources: sources}, nil
}

type telegrafProvider struct {
	sources []metrics.MetricsSource
}

func (p telegrafProvider) GetMetricsSources() []metrics.MetricsSource {
	return p.sources
}

func (p telegrafProvider) Name() string {
	return "Telegraf Source"
}

type telegrafPluginSource struct {
	name   string
	source string
	prefix string
	plugin telegraf.Input
}

func newTelegrafPluginSource(name string, plugin telegraf.Input, prefix string) *telegrafPluginSource {
	hostname := os.Getenv("POD_NODE_NAME")
	tsp := &telegrafPluginSource{name: name, plugin: plugin, source: hostname, prefix: prefix}
	return tsp
}

func (t *telegrafPluginSource) Name() string {
	return t.name + " plugin Source"
}

func (t *telegrafPluginSource) ScrapeMetrics(start, end time.Time) (*metrics.DataBatch, error) {
	result := &telegrafDataBatch{
		DataBatch: metrics.DataBatch{Timestamp: time.Now()},
		source:    t,
	}

	err := t.plugin.Gather(result)
	if err != nil {
		glog.Errorf("Error on Gather - plugin '%s' - error: %v", t.name, err)
	}

	glog.Infof("[ScrapeMetrics] plugin: '%v' metrics: '%v'", t.name, len(result.MetricPoints))
	return &result.DataBatch, nil
}

type telegrafDataBatch struct {
	metrics.DataBatch
	source *telegrafPluginSource
}

func (t *telegrafDataBatch) preparePoints(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	var ts time.Time
	if len(timestamp) > 0 {
		ts = timestamp[0]
	} else {
		ts = time.Now()
	}

	for metric, v := range fields {
		var value float64
		var err error
		switch v.(type) {
		case string:
			continue
		default:
			value, err = getFloat(v)
			if err != nil {
				glog.Errorf("Error, unsupported type '%v' - plugin: '%v' - metric: '%v' - value: '%v' - error: '%v'", reflect.TypeOf(v), t.source.name, metric, v, err)
				continue
			}
		}

		metricName := measurement + "." + metric
		metricName = strings.Replace(metricName, "_", ".", -1)
		if len(t.source.prefix) > 0 {
			metricName = t.source.prefix + "." + metricName
		}

		point := &metrics.MetricPoint{
			Metric:    metricName,
			Value:     value,
			Timestamp: ts.UnixNano() / 1000,
			Source:    t.source.source,
			Tags:      tags,
		}
		t.MetricPoints = append(t.MetricPoints, point)
	}
}

// AddFields adds a metric to the accumulator with the given measurement
// name, fields, and tags (and timestamp). If a timestamp is not provided,
// then the accumulator sets it to "now".
func (t *telegrafDataBatch) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddGauge is the same as AddFields, but will add the metric as a "Gauge" type
func (t *telegrafDataBatch) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddCounter is the same as AddFields, but will add the metric as a "Counter" type
func (t *telegrafDataBatch) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	t.preparePoints(measurement, fields, tags, timestamp...)
}

// AddSummary is the same as AddFields, but will add the metric as a "Summary" type
func (t *telegrafDataBatch) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddHistogram is the same as AddFields, but will add the metric as a "Histogram" type
func (t *telegrafDataBatch) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, timestamp ...time.Time) {
	glog.Fatal("not supported")
}

// AddMetric adds an metric to the accumulator.
func (t *telegrafDataBatch) AddMetric(telegraf.Metric) {
	glog.Fatal("not supported")
}

// SetPrecision sets the timestamp rounding precision.  All metrics addeds
// added to the accumulator will have their timestamp rounded to the
// nearest multiple of precision.
func (t *telegrafDataBatch) SetPrecision(precision time.Duration) {
	glog.Fatal("not supported")
}

// Report an error.
func (t *telegrafDataBatch) AddError(err error) {
	glog.Fatal("not supported")
}

// Upgrade to a TrackingAccumulator with space for maxTracked
// metrics/batches.
func (t *telegrafDataBatch) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	glog.Fatal("not supported")
	return nil
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
