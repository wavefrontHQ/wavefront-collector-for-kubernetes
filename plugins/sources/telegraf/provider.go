package telegraf

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/influxdata/telegraf"
	telegrafPlugins "github.com/influxdata/telegraf/plugins/inputs"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	wfTelegraf "github.com/wavefronthq/wavefront-kubernetes-collector/internal/telegraf"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf/redis"
)

type telegrafPluginSource struct {
	name    string
	source  string
	prefix  string
	tags    map[string]string
	plugin  telegraf.Input
	filters filter.Filter
}

func newTelegrafPluginSource(name string, plugin telegraf.Input, prefix string, tags map[string]string, filters filter.Filter) *telegrafPluginSource {
	tsp := &telegrafPluginSource{
		name:    name + "_plugin",
		plugin:  plugin,
		source:  util.GetNodeName(),
		prefix:  prefix,
		tags:    tags,
		filters: filters,
	}
	return tsp
}

func (t *telegrafPluginSource) Name() string {
	return "telegraf_" + t.name + "_source"
}

func (t *telegrafPluginSource) ScrapeMetrics(start, end time.Time) (*metrics.DataBatch, error) {
	result := &telegrafDataBatch{
		DataBatch: metrics.DataBatch{Timestamp: time.Now()},
		source:    t,
	}

	// Gather invokes callbacks on telegrafDataBatch
	err := t.plugin.Gather(result)
	if err != nil {
		glog.Errorf("error gathering %s metrics. error: %v", t.name, err)
	}
	glog.Infof("%s metrics: %d", t.name, len(result.MetricPoints))
	return &result.DataBatch, nil
}

// Telegraf provider
type telegrafProvider struct {
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

// NewProvider creates a Telegraf source
func NewProvider(uri *url.URL) (metrics.MetricsSourceProvider, error) {
	//for _, pair := range os.Environ() {
	//	glog.V(4).Infof("env: %v", pair)
	//}
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
		plugins = []string{"mem", "net", "netstat", "linux_sysctl_fs", "swap", "cpu", "disk", "diskio", "system", "kernel", "processes"}
	}

	filters := filter.FromQuery(vals)
	tags := flags.DecodeTags(vals)

	var sources []metrics.MetricsSource
	for _, name := range plugins {
		creator := telegrafPlugins.Inputs[strings.Trim(name, " ")]
		if creator != nil {
			plugin := creator()
			if handler, ok := handlers[name]; ok {
				err := handler.Init(plugin, vals)
				if err != nil {
					// bail if the plugin has special handlers
					return nil, err
				}
			}
			sources = append(sources, newTelegrafPluginSource(name, plugin, prefix, tags, filters))
		} else {
			glog.Errorf("telegraf plugin %s not found", name)
			var availablePlugins []string
			for name := range telegrafPlugins.Inputs {
				availablePlugins = append(availablePlugins, name)
			}
			glog.Infof("available telegraf plugins: '%v'", availablePlugins)
		}
	}

	name := ""
	if len(vals["name"]) > 0 {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["name"][0])
	}
	if name == "" {
		name = fmt.Sprintf("%s: default", ProviderName)
	}

	return &telegrafProvider{
		name:    name,
		sources: sources,
	}, nil
}

var handlers map[string]wfTelegraf.PluginHandler

func init() {
	handlers = make(map[string]wfTelegraf.PluginHandler)
	handlers["redis"] = redis.NewPluginHandler()
}
