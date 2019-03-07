package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
)

type CollectorRunOptions struct {
	MetricResolution      time.Duration
	MaxProcs              int
	Sources               flags.Uris
	Sinks                 flags.Uris
	Version               bool
	LabelSeparator        string
	IgnoredLabels         []string
	StoredLabels          []string
	SinkExportDataTimeout time.Duration
	EnableDiscovery       bool
	DiscoveryConfigFile   string
	InternalStatsPrefix   string
}

func NewCollectorRunOptions() *CollectorRunOptions {
	return &CollectorRunOptions{}
}

func (h *CollectorRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.Var(&h.Sources, "source", "source(s) to watch")
	fs.Var(&h.Sinks, "sink", "external sink(s) that receive data")
	fs.DurationVar(&h.MetricResolution, "metric_resolution", 60*time.Second, "The resolution at which the collector will retain metrics.")

	fs.IntVar(&h.MaxProcs, "max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	fs.BoolVar(&h.Version, "version", false, "print version info and exit")
	fs.StringVar(&h.LabelSeparator, "label_separator", ",", "separator used for joining labels")
	fs.StringSliceVar(&h.IgnoredLabels, "ignore_label", []string{}, "ignore this label when joining labels")
	fs.StringSliceVar(&h.StoredLabels, "store_label", []string{}, "store this label separately from joined labels with the same name (name) or with different name (newName=name)")
	fs.DurationVar(&h.SinkExportDataTimeout, "sink_export_data_timeout", 20*time.Second, "Timeout for exporting data to a sink")
	fs.BoolVar(&h.EnableDiscovery, "enable-discovery", true, "enable auto discovery")
	fs.StringVar(&h.DiscoveryConfigFile, "discovery_config", "", "optional discovery configuration file")
	fs.StringVar(&h.InternalStatsPrefix, "internal_stats_prefix", "kubernetes.", "optional prefix for internal collector stats")
}
