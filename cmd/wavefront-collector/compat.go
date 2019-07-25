package main

import (
	"net/url"
	"strings"

	"github.com/golang/glog"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
)

func handleBackwardsCompatibility(opt *options.CollectorRunOptions) {
	addInternalStatsSource(opt)
	cleanupSinkPrefix(opt)
}

// backwards compatibility: internal stats used to be included by default. It's now config driven.
func addInternalStatsSource(opt *options.CollectorRunOptions) {
	values := url.Values{}
	values.Add("prefix", opt.InternalStatsPrefix)

	u, err := url.Parse("?")
	if err != nil {
		glog.Errorf("error adding internal source: %v", err)
		return
	}
	u.RawQuery = values.Encode()
	opt.Sources = append(opt.Sources, flags.Uri{Key: "internal_stats", Val: *u})
}

// backwards compatibility: Sink level prefix used to only apply towards kubernetes.summary_api metrics
// this is now set on the kubernetes source. sink level prefixes now apply globally.
func cleanupSinkPrefix(opt *options.CollectorRunOptions) {
	if len(opt.Sinks) == 0 {
		glog.Fatalf("no sink configured")
	}
	sink := opt.Sinks[0]
	prefix := flags.DecodeValue(sink.Val.Query(), "prefix")
	if prefix != "" {
		// remove it from the sink
		opt.Sinks[0] = removeQueryKey(sink, "prefix")

		// set it on the kubernetes source
		for i, uri := range opt.Sources {
			if strings.SplitN(uri.Key, ".", 2)[0] == "kubernetes" {
				opt.Sources[i] = addQueryKey(uri, "prefix", prefix)
				return
			}
		}
	}
}

func addQueryKey(uri flags.Uri, key, value string) flags.Uri {
	val := &uri.Val
	values, err := url.ParseQuery(val.RawQuery)
	if err != nil {
		glog.Fatalf("error adding key: %v", err)
	}
	values.Add(key, value)
	val.RawQuery = values.Encode()
	return flags.Uri{Key: uri.Key, Val: *val}
}

func removeQueryKey(uri flags.Uri, key string) flags.Uri {
	val := &uri.Val
	values, err := url.ParseQuery(val.RawQuery)
	if err != nil {
		glog.Fatalf("error removing key :%v", err)
	}
	values.Del(key)
	val.RawQuery = values.Encode()
	return flags.Uri{Key: uri.Key, Val: *val}
}

// backwards compatibility: clusterName used to be specified on the sink. It's now a top-level config.
func resolveClusterName(name string, opt *options.CollectorRunOptions) string {
	if name == "" {
		sinkUrl, err := getWavefrontAddress(opt.Sinks)
		if err != nil {
			glog.Fatalf("Failed to get wavefront sink address: %v", err)
		}
		name = flags.DecodeValue(sinkUrl.Query(), "clusterName")
	}
	return name
}

// backwards compatibility: discovery config used to be a separate file. Now part of main config file.
func loadDiscoveryFileOrDie(file string) []discovery.PluginConfig {
	cfg, err := discovery.FromFile(file)
	if err != nil {
		glog.Fatalf("error loading discovery configuration: %v", err)
	}
	convertPromToPlugin(cfg)
	return cfg.PluginConfigs
}

func convertPromToPlugin(cfg *discovery.Config) {
	// convert PrometheusConfigs to PluginConfigs
	if len(cfg.PromConfigs) > 0 {
		glog.Warningf("Warning: PrometheusConfig has been deprecated. Use PluginConfig.")
		toAppend := make([]discovery.PluginConfig, len(cfg.PromConfigs))
		for i, promCfg := range cfg.PromConfigs {
			toAppend[i] = discovery.PluginConfig{
				Name:          promCfg.Name,
				Type:          "prometheus",
				Port:          promCfg.Port,
				Scheme:        promCfg.Scheme,
				Path:          promCfg.Path,
				Source:        promCfg.Source,
				Prefix:        promCfg.Prefix,
				Tags:          promCfg.Tags,
				IncludeLabels: promCfg.IncludeLabels,
				Filters:       promCfg.Filters,
				Selectors: discovery.Selectors{
					ResourceType: promCfg.ResourceType,
					Namespaces:   []string{promCfg.Namespace},
				},
			}
			labels := map[string][]string{}
			for k, v := range promCfg.Labels {
				labels[k] = []string{v}
			}
			toAppend[i].Selectors.Labels = labels
		}
		cfg.PluginConfigs = append(cfg.PluginConfigs, toAppend...)
	}
}
