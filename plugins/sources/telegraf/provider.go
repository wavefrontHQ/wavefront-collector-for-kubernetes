package telegraf

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	telegrafPlugins "github.com/influxdata/telegraf/plugins/inputs"
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

type telegrafPluginSource struct {
	name    string
	source  string
	prefix  string
	tags    map[string]string
	plugin  telegraf.Input
	filters filter.Filter

	pointsCollected gm.Counter
	pointsFiltered  gm.Counter
	errors          gm.Counter
	targetPPS       gm.Counter
	targetEPS       gm.Counter
}

func newTelegrafPluginSource(name string, plugin telegraf.Input, prefix string, tags map[string]string, filters filter.Filter, discovered string) *telegrafPluginSource {
	pt := map[string]string{"type": "telegraf." + name}
	collected := reporting.EncodeKey("source.points.collected", pt)
	filtered := reporting.EncodeKey("source.points.filtered", pt)
	errors := reporting.EncodeKey("source.collect.errors", pt)

	tsp := &telegrafPluginSource{
		name:            name + "_plugin",
		plugin:          plugin,
		source:          util.GetNodeName(),
		prefix:          prefix,
		tags:            tags,
		filters:         filters,
		pointsCollected: gm.GetOrRegisterCounter(collected, gm.DefaultRegistry),
		pointsFiltered:  gm.GetOrRegisterCounter(filtered, gm.DefaultRegistry),
		errors:          gm.GetOrRegisterCounter(errors, gm.DefaultRegistry),
	}
	if discovered != "" {
		pt = extractTags(tags, name, discovered)
		tsp.targetPPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.points.collected", pt), gm.DefaultRegistry)
		tsp.targetEPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.collect.errors", pt), gm.DefaultRegistry)
	}
	return tsp
}

func extractTags(tags map[string]string, name, discovered string) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if k == "pod" || k == "service" || k == "namespace" {
			result[k] = v
		}
	}
	if discovered != "" {
		result["discovered"] = discovered
	}
	result["type"] = "telegraf." + name
	return result
}

func (t *telegrafPluginSource) Name() string {
	return "telegraf_" + t.name + "_source"
}

func (t *telegrafPluginSource) ScrapeMetrics() (*metrics.DataBatch, error) {
	result := &telegrafDataBatch{
		DataBatch: metrics.DataBatch{Timestamp: time.Now()},
		source:    t,
	}

	// Gather invokes callbacks on telegrafDataBatch
	err := t.plugin.Gather(result)
	if err != nil {
		t.errors.Inc(1)
		if t.targetEPS != nil {
			t.targetEPS.Inc(1)
		}
		log.Errorf("error gathering %s metrics. error: %v", t.name, err)
	}
	count := len(result.MetricPoints)
	log.Infof("%s metrics: %d", t.Name(), count)
	t.pointsCollected.Inc(int64(count))
	if t.targetPPS != nil {
		t.targetPPS.Inc(int64(count))
	}
	return &result.DataBatch, nil
}

// Telegraf provider
type telegrafProvider struct {
	metrics.DefaultMetricsSourceProvider
	name    string
	sources []metrics.MetricsSource
}

func (p telegrafProvider) GetMetricsSources() []metrics.MetricsSource {
	return p.sources
}

func (p telegrafProvider) Name() string {
	return p.name
}

const ProviderName = "telegraf_provider"

var defaultPlugins = []string{"mem", "net", "netstat", "linux_sysctl_fs", "swap", "cpu", "disk", "diskio", "system", "kernel", "processes"}

// NewProvider creates a Telegraf source
func NewProvider(uri *url.URL) (metrics.MetricsSourceProvider, error) {
	vals := uri.Query()

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
		plugins = defaultPlugins
	}

	filters := filter.FromQuery(vals)
	tags := flags.DecodeTags(vals)
	discovered := flags.DecodeValue(vals, "discovered")

	var sources []metrics.MetricsSource
	for _, name := range plugins {
		creator := telegrafPlugins.Inputs[strings.Trim(name, " ")]
		if creator != nil {
			plugin := creator()
			if discovered != "" {
				err := initPlugin(plugin, vals)
				if err != nil {
					// bail if discovered and error initializing
					log.Errorf("error creating plugin: %s err: %s", name, err)
					return nil, err
				}
			}
			sources = append(sources, newTelegrafPluginSource(name, plugin, prefix, tags, filters, discovered))
		} else {
			log.Errorf("telegraf plugin %s not found", name)
			var availablePlugins []string
			for name := range telegrafPlugins.Inputs {
				availablePlugins = append(availablePlugins, name)
			}
			log.Infof("available telegraf plugins: '%v'", availablePlugins)
		}
	}

	name := ""
	if len(vals["name"]) > 0 {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["name"][0])
	}
	if name == "" {
		name = fmt.Sprintf("%s: %v", ProviderName, plugins)
	}

	return &telegrafProvider{
		name:    name,
		sources: sources,
	}, nil
}
