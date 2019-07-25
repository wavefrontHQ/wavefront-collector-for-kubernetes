package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
)

type CollectorRunOptions struct {
	// supported flags
	Version         bool
	EnableProfiling bool
	Daemon          bool
	ConfigFile      string

	// deprecated flags
	MetricResolution      time.Duration
	MaxProcs              int
	Sources               flags.Uris
	Sinks                 flags.Uris
	SinkExportDataTimeout time.Duration
	EnableDiscovery       bool
	DiscoveryConfigFile   string
	InternalStatsPrefix   string
	ScrapeTimeout         time.Duration
	logLevel              int
}

func NewCollectorRunOptions() *CollectorRunOptions {
	return &CollectorRunOptions{}
}

func (h *CollectorRunOptions) AddFlags(fs *pflag.FlagSet) {
	// supported flags
	fs.BoolVar(&h.Version, "version", false, "print version info and exit")
	fs.BoolVar(&h.EnableProfiling, "profile", false, "enable pprof")
	fs.BoolVar(&h.Daemon, "daemon", false, "enable daemon mode")
	fs.StringVar(&h.ConfigFile, "config_file", "", "required configuration file")

	// deprecated flags
	fs.DurationVar(&h.MetricResolution, "metric_resolution", 60*time.Second, "The resolution at which the collector will collect metrics")
	fs.MarkDeprecated("metric_resolution", "set defaultCollectionInterval in configuration file")
	fs.IntVar(&h.MaxProcs, "max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	fs.MarkDeprecated("max_procs", "set maxProcs in configuration file")
	fs.Var(&h.Sources, "source", "source(s) to watch")
	fs.MarkDeprecated("source", "set sources in configuration file")
	fs.Var(&h.Sinks, "sink", "external sink(s) that receive data")
	fs.MarkDeprecated("sink", "set sinks in configuration file")
	fs.DurationVar(&h.SinkExportDataTimeout, "sink_export_data_timeout", 20*time.Second, "Timeout for exporting data to a sink")
	fs.MarkDeprecated("sink_export_data_timeout", "set sinkExportDataTimeout in configuration file")
	fs.BoolVar(&h.EnableDiscovery, "enable-discovery", true, "enable auto discovery")
	fs.MarkDeprecated("enable-discovery", "set enableDiscovery in configuration file")
	fs.StringVar(&h.DiscoveryConfigFile, "discovery_config", "", "optional discovery configuration file")
	fs.MarkDeprecated("discovery_config", "set discovery_configs in configuration file")
	fs.StringVar(&h.InternalStatsPrefix, "internal_stats_prefix", "kubernetes.", "optional prefix for internal collector stats")
	fs.MarkDeprecated("internal_stats_prefix", "set internal_stats_source in configuration file")
	fs.DurationVar(&h.ScrapeTimeout, "scrape_timeout", 20*time.Second, "The per-source scrape timeout")
	fs.MarkDeprecated("scrape_timeout", "set in configuration file")
	fs.IntVar(&h.logLevel, "v", 2, "log level for V logs")
	fs.MarkDeprecated("v", "flag is ignored")
}
