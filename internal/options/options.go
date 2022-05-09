// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/flags"
)

type CollectorRunOptions struct {
	// supported flags
	Version         bool
	EnableProfiling bool
	Daemon          bool
	ScrapeCluster   bool
	ScrapeNodes     string
	ConfigFile      string
	LogLevel        string
	MaxProcs        int

	// An experimental flag for forcing a garbage collection and releasing memory.
	// See https://utcc.utoronto.ca/~cks/space/blog/programming/GoNoMemoryFreeing for reference.
	// Basically Go holds on to more memory than is necessary resulting in larger heap usage.
	// Enabling this flag causes the collector to call debug.FreeOSMemory after every sink.send() call.
	// Go 1.13 showed a 30% lower memory usage vs Go 1.12. Enabling this flag was observed to result in a further ~30%
	// reduction in memory usage but the impact of forcing a GC run so frequently has not been thoroughly tested.
	ForceGC bool

	// deprecated flags
	MetricResolution      time.Duration
	Sources               flags.Uris
	Sinks                 flags.Uris
	SinkExportDataTimeout time.Duration
	EnableDiscovery       bool
	EnableRuntimeConfigs  bool
	DiscoveryConfigFile   string
	InternalStatsPrefix   string
	ScrapeTimeout         time.Duration
	logLevel              int
}

func NewCollectorRunOptions() *CollectorRunOptions {
	return &CollectorRunOptions{}
}

func (opts *CollectorRunOptions) Parse(fs *pflag.FlagSet, args []string) error {
	// supported flags
	fs.BoolVar(&opts.Version, "version", false, "print version info and exit")
	fs.BoolVar(&opts.EnableProfiling, "profile", false, "enable pprof")
	fs.BoolVar(&opts.Daemon, "daemon", false, "enable daemon mode")
	fs.BoolVar(&opts.ScrapeCluster, "scrape-cluster", true, "whether to participate in scraping cluster metrics (uses leader election)")
	fs.StringVar(&opts.ScrapeNodes, "scrape-nodes", "all", "which nodes to scrape (all, own, node)")
	fs.StringVar(&opts.ConfigFile, "config-file", "", "required configuration file")
	fs.StringVar(&opts.LogLevel, "log-level", "info", "one of info, debug or trace")
	fs.IntVar(&opts.MaxProcs, "max-procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	fs.BoolVar(&opts.ForceGC, "force-gc", false, "experimental flag that periodically forces the release of unused memory")

	// deprecated flags
	fs.DurationVar(&opts.MetricResolution, "metric-resolution", 60*time.Second, "The resolution at which the collector will collect metrics")
	fs.MarkDeprecated("metric-resolution", "set defaultCollectionInterval in configuration file")
	fs.Var(&opts.Sources, "source", "source(s) to watch")
	fs.MarkDeprecated("source", "set sources in configuration file")
	fs.Var(&opts.Sinks, "sink", "external sink(s) that receive data")
	fs.MarkDeprecated("sink", "set sinks in configuration file")
	fs.DurationVar(&opts.SinkExportDataTimeout, "sink-export-data-timeout", 20*time.Second, "Timeout for exporting data to a sink")
	fs.MarkDeprecated("sink-export-data-timeout", "set sinkExportDataTimeout in configuration file")
	fs.BoolVar(&opts.EnableDiscovery, "enable-discovery", true, "enable auto discovery")
	fs.MarkDeprecated("enable-discovery", "set enableDiscovery in configuration file")
	fs.BoolVar(&opts.EnableRuntimeConfigs, "enable-runtime-configs", false, "enable runtime configs")
	fs.MarkDeprecated("enable-runtime-configs", "set enable-runtime-configs in configuration file")
	fs.StringVar(&opts.DiscoveryConfigFile, "discovery-config", "", "optional discovery configuration file")
	fs.MarkDeprecated("discovery-config", "set discovery_configs in configuration file")
	fs.StringVar(&opts.InternalStatsPrefix, "internal-stats-prefix", "kubernetes.", "optional prefix for internal collector stats")
	fs.MarkDeprecated("internal-stats-prefix", "set internal_stats_source in configuration file")
	fs.DurationVar(&opts.ScrapeTimeout, "scrape-timeout", 20*time.Second, "The per-source scrape timeout")
	fs.MarkDeprecated("scrape-timeout", "set in configuration file")
	fs.IntVar(&opts.logLevel, "v", 2, "log level for V logs")
	fs.MarkDeprecated("v", "use log-level instead")

	if err := fs.Parse(args); err != nil {
		return err
	}

    if err := opts.verifyFlagCombos(fs); err != nil {
        return err
    }

	if opts.Daemon {
		opts.ScrapeCluster = true
		opts.ScrapeNodes = "own"
	}

    return nil
}

func (opts *CollectorRunOptions) verifyFlagCombos(fs *pflag.FlagSet) error {
    var daemonSpecified bool
    var scrapeNodesSpecified bool
    var scrapeClusterSpecified bool
    fs.Visit(func(flag *pflag.Flag) {
        switch flag.Name {
        case "daemon":
            daemonSpecified = flag.Changed
        case "scrape-nodes":
            scrapeNodesSpecified = flag.Changed
        case "scrape-cluster":
            scrapeClusterSpecified = flag.Changed
        }
    })

    if daemonSpecified && (scrapeNodesSpecified || scrapeClusterSpecified) {
        return errors.New("cannot set daemon with either scrape-nodes or scrape-cluster")
    }
    return nil
}

func Parse() *CollectorRunOptions {
	opts := NewCollectorRunOptions()
	fs := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	if err := opts.Parse(fs, os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return opts
}
