// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/flags"
)

// Convert converts options into a configuration instance for backwards compatibility
func (opts *CollectorRunOptions) Convert() (*configuration.Config, error) {
	cfg := &configuration.Config{}
	cfg.DefaultCollectionInterval = opts.MetricResolution
	cfg.FlushInterval = opts.MetricResolution
	cfg.SinkExportDataTimeout = opts.SinkExportDataTimeout
	cfg.EnableDiscovery = opts.EnableDiscovery
	cfg.Daemon = opts.Daemon

	if len(opts.Sources) == 0 {
		return nil, fmt.Errorf("missing sources")
	}
	if len(opts.Sinks) == 0 {
		return nil, fmt.Errorf("missing sink")
	}

	addSources(cfg, opts.Sources)
	addSinks(cfg, opts.Sinks)
	addInternalStatsSource(cfg, opts.InternalStatsPrefix)
	extractSinkProperties(cfg)

	if opts.EnableDiscovery {
		cfg.DiscoveryConfig.EnableRuntimePlugins = opts.EnableRuntimeConfigs
		if cfg.DiscoveryConfig.DiscoveryInterval == 0 {
			cfg.DiscoveryConfig.DiscoveryInterval = 5 * time.Minute
		}
		if opts.DiscoveryConfigFile != "" {
			cfg.DiscoveryConfig.PluginConfigs = loadDiscoveryFileOrDie(opts.DiscoveryConfigFile)
		}
	}
	return cfg, nil
}

// backwards compatibility: discovery config used to be a separate file. Now part of main config file.
func loadDiscoveryFileOrDie(file string) []discovery.PluginConfig {
	cfg, err := discovery.FromFile(file)
	if err != nil {
		log.Fatalf("error loading discovery configuration: %v", err)
	}
	discovery.ConvertPromToPlugin(cfg)
	return cfg.PluginConfigs
}

// backwards compatibility:
// 1. Sink level prefix used to only apply towards kubernetes.summary_api metrics
// this is now set on the kubernetes source. sink level prefixes now apply globally.
// 2. clusterName used to be specified on the sink. Now a top level property.
func extractSinkProperties(cfg *configuration.Config) {
	if len(cfg.Sinks) == 0 {
		log.Fatalf("no sink configured")
	}
	sink := cfg.Sinks[0]
	prefix := sink.Prefix
	if prefix != "" {
		// remove it from the sink
		sink.Prefix = ""

		// set it on the kubernetes source
		cfg.Sources.SummaryConfig.Prefix = prefix
	}
	cfg.ClusterName = sink.ClusterName
}

func addSources(cfg *configuration.Config, sources flags.Uris) {
	cfg.Sources = &configuration.SourceConfig{}
	for _, src := range sources {
		switch src.Key {
		case "kubernetes.summary_api":
			addSummarySource(cfg, src)
		case "kubernetes.state":
			addStateSource(cfg, src)
		case "prometheus":
			addPrometheusSource(cfg, src)
		case "telegraf":
			addTelegrafSource(cfg, src)
		case "systemd":
			addSystemdSource(cfg, src)
		default:
			log.Errorf("invalid source: %s", src.Key)
			return
		}
	}
}

func addSummarySource(cfg *configuration.Config, uri flags.Uri) {
	vals := uri.Val.Query()
	summary := &configuration.SummarySourceConfig{
		KubeletPort:       flags.DecodeValue(vals, "kubeletPort"),
		KubeletHttps:      flags.DecodeValue(vals, "kubeletHttps"),
		InClusterConfig:   flags.DecodeValue(vals, "inClusterConfig"),
		UseServiceAccount: flags.DecodeValue(vals, "useServiceAccount"),
		Insecure:          flags.DecodeValue(vals, "insecure"),
		Auth:              flags.DecodeValue(vals, "auth"),
		Transforms:        getTransforms(vals),
	}
	if uri.Val.Scheme != "" {
		u := uri.Val
		summary.URL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	}
	cfg.Sources.SummaryConfig = summary
}

func addStateSource(cfg *configuration.Config, uri flags.Uri) {
	vals := uri.Val.Query()
	state := &configuration.KubernetesStateSourceConfig{
		Transforms: getTransforms(vals),
	}
	cfg.Sources.StateConfig = state
}

func addPrometheusSource(cfg *configuration.Config, uri flags.Uri) {
	vals := uri.Val.Query()
	prom := &configuration.PrometheusSourceConfig{
		URL:        flags.DecodeValue(vals, "url"),
		Transforms: getTransforms(vals),
	}
	cfg.Sources.PrometheusConfigs = append(cfg.Sources.PrometheusConfigs, prom)
}

func addTelegrafSource(cfg *configuration.Config, uri flags.Uri) {
	vals := uri.Val.Query()
	tel := &configuration.TelegrafSourceConfig{
		Transforms: getTransforms(vals),
	}
	plugins := flags.DecodeValue(vals, "plugins")
	if len(plugins) > 0 {
		tel.Plugins = strings.Split(plugins, ",")
	}
	cfg.Sources.TelegrafConfigs = append(cfg.Sources.TelegrafConfigs, tel)
}

func addSystemdSource(cfg *configuration.Config, uri flags.Uri) {
	vals := uri.Val.Query()
	systemd := &configuration.SystemdSourceConfig{
		IncludeTaskMetrics:      flags.DecodeBoolean(vals, "taskMetrics"),
		IncludeRestartMetrics:   flags.DecodeBoolean(vals, "restartMetrics"),
		IncludeStartTimeMetrics: flags.DecodeBoolean(vals, "startTimeMetrics"),
		UnitWhitelist:           vals["unitWhitelist"],
		UnitBlacklist:           vals["unitBlacklist"],
		Transforms:              getTransforms(vals),
	}
	cfg.Sources.SystemdConfig = systemd
}

func addInternalStatsSource(cfg *configuration.Config, prefix string) {
	cfg.Sources.StatsConfig = &configuration.StatsSourceConfig{
		Transforms: configuration.Transforms{
			Prefix: prefix,
		},
	}
}

func addSinks(cfg *configuration.Config, sinks flags.Uris) {
	for _, sink := range sinks {
		switch sink.Key {
		case "wavefront":
			addWavefrontSink(cfg, sink)
		default:
			return
		}
	}
}

func addWavefrontSink(cfg *configuration.Config, uri flags.Uri) {
	sink := &configuration.WavefrontSinkConfig{}
	vals := uri.Val.Query()

	sink.Server = flags.DecodeValue(vals, "server")
	sink.Token = flags.DecodeValue(vals, "token")
	sink.ProxyAddress = flags.DecodeValue(vals, "proxyAddress")
	sink.ClusterName = flags.DecodeValue(vals, "clusterName")
	sink.Transforms = getTransforms(vals)

	cfg.Sinks = append(cfg.Sinks, sink)
}

func getTransforms(vals map[string][]string) configuration.Transforms {
	return configuration.Transforms{
		Prefix:  flags.DecodeValue(vals, "prefix"),
		Source:  flags.DecodeValue(vals, "source"),
		Tags:    flags.DecodeTags(vals),
		Filters: filter.FromQuery(vals),
	}
}
