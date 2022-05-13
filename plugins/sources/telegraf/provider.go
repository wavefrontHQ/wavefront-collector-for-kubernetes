// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"fmt"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"

	"github.com/influxdata/telegraf"
	telegrafPlugins "github.com/influxdata/telegraf/plugins/inputs"
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

type telegrafPluginSource struct {
	name           string
	source         string
	prefix         string
	tags           map[string]string
	plugin         telegraf.Input
	pluginPrefix   string
	filters        filter.Filter
	autoDiscovered bool

	pointsCollected gm.Counter
	pointsFiltered  gm.Counter
	errors          gm.Counter

	targetTags map[string]string
	targetPPS  gm.Counter
	targetEPS  gm.Counter
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
		autoDiscovered:  len(discovered) > 0,
		pointsCollected: gm.GetOrRegisterCounter(collected, gm.DefaultRegistry),
		pointsFiltered:  gm.GetOrRegisterCounter(filtered, gm.DefaultRegistry),
		errors:          gm.GetOrRegisterCounter(errors, gm.DefaultRegistry),
	}
	if discovered != "" {
		tsp.targetTags = extractTags(tags, pluginType, discovered)
		tsp.targetPPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.points.collected", tsp.targetTags), gm.DefaultRegistry)
		tsp.targetEPS = gm.GetOrRegisterCounter(reporting.EncodeKey("target.collect.errors", tsp.targetTags), gm.DefaultRegistry)
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

func (t *telegrafPluginSource) Cleanup() {
	gm.Unregister(reporting.EncodeKey("target.collect.errors", t.targetTags))
	gm.Unregister(reporting.EncodeKey("target.collect.errors", t.targetTags))
}

func (t *telegrafPluginSource) AutoDiscovered() bool {
	return t.autoDiscovered
}

func (t *telegrafPluginSource) Name() string {
	return "telegraf_" + t.name + "_source"
}

func (t *telegrafPluginSource) Scrape() (*metrics.Batch, error) {
	result := &telegrafDataBatch{
		Batch:  metrics.Batch{Timestamp: time.Now()},
		source: t,
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
	count := len(result.Points)

	log.WithFields(log.Fields{
		"name":          t.Name(),
		"total_metrics": count,
	}).Debug("Scraping completed")

	t.pointsCollected.Inc(int64(count))
	if t.targetPPS != nil {
		t.targetPPS.Inc(int64(count))
	}
	return &result.Batch, nil
}

// Telegraf provider
type telegrafProvider struct {
	metrics.DefaultSourceProvider
	name              string
	useLeaderElection bool
	sources           []metrics.Source
}

func (p telegrafProvider) GetMetricsSources() []metrics.Source {
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
func NewProvider(cfg configuration.TelegrafSourceConfig) (metrics.SourceProvider, error) {
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
	hostPlugin := true

	var sources []metrics.Source
	for _, name := range plugins {
		if !util.ShouldScrapeNodeMetrics() && pluginType(name) == "telegraf_host" {
			continue
		}
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
			hostPlugin = hostPlugin && (pluginType(name) == "telegraf_host")
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
	useLeaderElection := cfg.UseLeaderElection || (cfg.Discovered == "" && !hostPlugin)

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
