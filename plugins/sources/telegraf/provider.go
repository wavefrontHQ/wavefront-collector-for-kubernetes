// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	telegrafPlugins "github.com/influxdata/telegraf/plugins/inputs"
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

type telegrafPluginSource struct {
	name         string
	source       string
	prefix       string
	tags         map[string]string
	plugin       telegraf.Input
	pluginPrefix string
	filters      filter.Filter

	pointsCollected gm.Counter
	pointsFiltered  gm.Counter
	errors          gm.Counter
	targetPPS       gm.Counter
	targetEPS       gm.Counter
}

func newTelegrafPluginSource(name string, plugin telegraf.Input, prefix string, tags map[string]string, filters filter.Filter, discovered string) *telegrafPluginSource {
	pluginType := pluginType(name) + "." + name
	pt := map[string]string{"type": pluginType}
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
		pt = extractTags(tags, pluginType, discovered)
		tsp.targetPPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.points.collected", pt), gm.DefaultRegistry)
		tsp.targetEPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.collect.errors", pt), gm.DefaultRegistry)
	}
	return tsp
}

func extractTags(tags map[string]string, name, discovered string) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if k == "pod" || k == "service" || k == "namespace" || k == "node" {
			result[k] = v
		}
	}
	if discovered != "" {
		result["discovered"] = discovered
	} else {
		result["discovered"] = "static"
	}
	result["type"] = name
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

	log.WithFields(log.Fields{
		"name":          t.Name(),
		"total_metrics": count,
	}).Debug("Scraping completed")

	t.pointsCollected.Inc(int64(count))
	if t.targetPPS != nil {
		t.targetPPS.Inc(int64(count))
	}
	return &result.DataBatch, nil
}

// Telegraf provider
type telegrafProvider struct {
	metrics.DefaultMetricsSourceProvider
	name              string
	useLeaderElection bool
	sources           []metrics.MetricsSource
}

func (p telegrafProvider) GetMetricsSources() []metrics.MetricsSource {
	// only the leader will collect from a static source (not auto-discovered) that is not a host plugin
	if p.useLeaderElection && !leadership.Leading() {
		log.Infof("not scraping sources from: %s. current leader: %s", p.name, leadership.Leader())
		return nil
	}
	return p.sources
}

func (p telegrafProvider) Name() string {
	return p.name
}

const providerName = "telegraf_provider"

var hostPlugins = []string{"mem", "net", "netstat", "linux_sysctl_fs", "swap", "cpu", "disk", "diskio", "system", "kernel", "processes"}

// NewProvider creates a Telegraf source
func NewProvider(cfg configuration.TelegrafSourceConfig) (metrics.MetricsSourceProvider, error) {
	prefix := configuration.GetStringValue(cfg.Prefix, "")
	if len(prefix) > 0 {
		prefix = strings.Trim(prefix, ".")
	}

	plugins := cfg.Plugins
	if len(plugins) == 0 {
		plugins = hostPlugins
	}

	filters := filter.FromConfig(cfg.Filters)
	tags := cfg.Tags
	discovered := cfg.Discovered

	var sources []metrics.MetricsSource
	for _, name := range plugins {
		creator := telegrafPlugins.Inputs[strings.Trim(name, " ")]
		if creator != nil {
			plugin := creator()
			if cfg.Conf != "" {
				err := initPlugin(plugin, cfg.Conf)
				if err != nil {
					return nil, fmt.Errorf("error creating plugin: %s err: %s", name, err)
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
			return nil, fmt.Errorf("telegraf plugin not found: %s", name)
		}
	}

	name := cfg.Name
	if len(name) > 0 {
		name = fmt.Sprintf("%s: %s", providerName, name)
	} else {
		name = fmt.Sprintf("%s: %v", providerName, plugins)
	}

	// use leader election if static source (not discovered) and is not a host plugin
	useLeaderElection := cfg.Discovered == "" && cfg.Conf != ""

	return &telegrafProvider{
		name:              name,
		useLeaderElection: useLeaderElection,
		sources:           sources,
	}, nil
}

func pluginType(plugin string) string {
	for _, name := range hostPlugins {
		if plugin == name {
			return "telegraf_host"
		}
	}
	return "telegraf"
}
