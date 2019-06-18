package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/manager"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
	kubeFlag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/klog"
)

var (
	version string
	commit  string
)

func main() {
	// Create go-kit logger (wrapper around glog)
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestamp)
	logger = level.NewFilter(logger, level.AllowDebug())

	// Overriding the default glog with our go-kit glog implementation.
	// Thus we need to pass it our go-kit logger object.
	glog.ClampLevel(6)
	glog.SetLogger(logger)

	klog.ClampLevel(6)
	klog.SetLogger(logger)

	opt := options.NewCollectorRunOptions()
	opt.AddFlags(pflag.CommandLine)
	kubeFlag.InitFlags()

	if opt.Version {
		fmt.Println(fmt.Sprintf("version: %s\ncommit: %s", version, commit))
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	sourceManager := createSourceManagerOrDie(opt.Sources, opt.InternalStatsPrefix, opt.ScrapeTimeout)
	sinkManager := createAndInitSinksOrDie(opt.Sinks, opt.SinkExportDataTimeout)

	man, err := manager.NewManager(sourceManager, nil, sinkManager,
		opt.MetricResolution, manager.DefaultScrapeOffset, manager.DefaultMaxParallelism)
	if err != nil {
		glog.Fatalf("Failed to create main manager: %v", err)
	}
	man.Start()
	waitForStop()
}

func createSourceManagerOrDie(src flags.Uris, statsPrefix string, scrapeTimeout time.Duration) metrics.MetricsSource {
	sourceFactory := sources.NewSourceFactory()
	sourceList := sourceFactory.BuildAll(src, statsPrefix)

	for _, source := range sourceList {
		glog.Infof("Starting with source %s", source.Name())
	}

	sourceManager, err := sources.NewSourceManager(sourceList, scrapeTimeout)
	if err != nil {
		glog.Fatalf("Failed to create source manager: %v", err)
	}
	return sourceManager
}

func createAndInitSinksOrDie(sinkAddresses flags.Uris, sinkExportDataTimeout time.Duration) metrics.DataSink {
	sinksFactory := sinks.NewSinkFactory()
	sinkList := sinksFactory.BuildAll(sinkAddresses)

	for _, sink := range sinkList {
		glog.Infof("Starting with %s", sink.Name())
	}
	sinkManager, err := sinks.NewDataSinkManager(sinkList, sinkExportDataTimeout, sinks.DefaultSinkStopTimeout)
	if err != nil {
		glog.Fatalf("Failed to create sink manager: %v", err)
	}
	return sinkManager
}

func waitForStop() {
	select {}
}
