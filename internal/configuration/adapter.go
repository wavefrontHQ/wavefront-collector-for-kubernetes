package configuration

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
)

type adapter interface {
	convert() (flags.Uri, error)
}

// Converts a configuration into option flags
func (c Config) Convert() (*options.CollectorRunOptions, error) {
	// For now we internally convert configs to flags/Uris to ease the transition from flags to configurations.
	// This code can be removed when sources, sinks and discovery code are wired to use configs instead of URLs.

	opts := options.NewCollectorRunOptions()
	opts.MetricResolution = c.CollectionInterval
	opts.SinkExportDataTimeout = c.SinkExportDataTimeout
	opts.ScrapeTimeout = c.ScrapeTimeout
	opts.MaxProcs = c.MaxProcs
	opts.EnableDiscovery = c.EnableDiscovery

	if err := addSource(c.SummaryConfig, opts); err != nil {
		return nil, err
	}
	if err := addSource(c.SystemdConfig, opts); err != nil {
		return nil, err
	}
	if err := addSource(c.StatsConfig, opts); err != nil {
		return nil, err
	}

	for _, cfg := range c.PrometheusConfigs {
		if err := addSource(cfg, opts); err != nil {
			return nil, err
		}
	}
	for _, cfg := range c.TelegrafConfigs {
		if err := addSource(cfg, opts); err != nil {
			return nil, err
		}
	}
	for _, cfg := range c.Sinks {
		cfg.ClusterName = c.ClusterName
		if err := addSink(cfg, opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

// converts a Wavefront sink configuration to a Uri format
func (w WavefrontSinkConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "server", w.Server)
	addVal(values, "token", w.Token)
	addVal(values, "proxyAddress", w.ProxyAddress)
	addVal(values, "prefix", w.Prefix)
	addVal(values, "testMode", strconv.FormatBool(w.TestMode))
	addVal(values, "clusterName", w.ClusterName)
	utils.EncodeFilters(values, w.Filters)
	utils.EncodeTags(values, "", w.Tags)
	return buildUri("wavefront", "", values.Encode())
}

// converts a Kubernetes summary source configuration to a Uri format
func (k SummaySourceConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "kubeletPort", k.KubeletPort)
	addVal(values, "kubeletHttps", k.KubeletHttps)
	addVal(values, "inClusterConfig", k.InClusterConfig)
	addVal(values, "useServiceAccount", k.UseServiceAccount)
	addVal(values, "insecure", k.Insecure)
	addVal(values, "auth", k.Auth)
	addVal(values, "prefix", k.Prefix)
	utils.EncodeFilters(values, k.Filters)
	utils.EncodeTags(values, "", k.Tags)
	return buildUri("kubernetes.summary_api", k.URL, values.Encode())
}

// converts a Prometheus source configuration to a Uri format
func (p PrometheusSourceConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "url", p.URL)
	addVal(values, "prefix", p.Prefix)
	addVal(values, "source", p.Source)
	utils.EncodeFilters(values, p.Filters)
	utils.EncodeTags(values, "", p.Tags)
	return buildUri("prometheus", "", values.Encode())
}

// converts a Telegraf source configuration to a Uri format
func (t TelegrafSourceConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "plugins", strings.Join(t.Plugins, ","))
	addVal(values, "prefix", t.Prefix)
	utils.EncodeFilters(values, t.Filters)
	utils.EncodeTags(values, "", t.Tags)
	return buildUri("telegraf", "", values.Encode())
}

// converts a Systemd source configuration to a Uri format
func (s SystemdSourceConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "prefix", s.Prefix)
	addVal(values, "taskMetrics", strconv.FormatBool(s.IncludeTaskMetrics))
	addVal(values, "startTimeMetrics", strconv.FormatBool(s.IncludeStartTimeMetrics))
	addVal(values, "restartMetrics", strconv.FormatBool(s.IncludeRestartMetrics))

	for _, val := range s.UnitWhitelist {
		addVal(values, "unitWhitelist", val)
	}
	for _, val := range s.UnitBlacklist {
		addVal(values, "unitBlacklist", val)
	}
	utils.EncodeFilters(values, s.Filters)
	utils.EncodeTags(values, "", s.Tags)
	return buildUri("systemd", "", values.Encode())
}

func buildUri(key, address, rawQuery string) (flags.Uri, error) {
	u, err := url.Parse(address + "?")
	if err != nil {
		return flags.Uri{}, err
	}
	u.RawQuery = rawQuery

	uri := flags.Uri{Key: key, Val: *u}
	return uri, nil
}

// converts an Internal stats source configuration to a Uri format
func (s StatsSourceConfig) convert() (flags.Uri, error) {
	values := url.Values{}
	addVal(values, "prefix", s.Prefix)
	utils.EncodeFilters(values, s.Filters)
	utils.EncodeTags(values, "", s.Tags)
	return buildUri("internal_stats", "", values.Encode())
}

func addVal(values url.Values, key, val string) {
	if val != "" {
		values.Add(key, val)
	}
}

func addSource(a adapter, opts *options.CollectorRunOptions) error {
	if a != nil && !reflect.ValueOf(a).IsNil() {
		opt, err := a.convert()
		if err != nil {
			return fmt.Errorf("error converting source: %v", err)
		}
		opts.Sources = append(opts.Sources, opt)
	}
	return nil
}

func addSink(a adapter, opts *options.CollectorRunOptions) error {
	if a != nil && !reflect.ValueOf(a).IsNil() {
		opt, err := a.convert()
		if err != nil {
			return fmt.Errorf("error converting sink: %v", err)
		}
		opts.Sinks = append(opts.Sinks, opt)
	}
	return nil
}
