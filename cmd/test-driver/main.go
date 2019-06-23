package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
	pluginsDiscovery "github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
	pluginsTelegraf "github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

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

	testDiscovery()

	//sourceManager := createSourceManagerOrDie(opt.Sources, opt.InternalStatsPrefix, opt.ScrapeTimeout)
	//sinkManager := createAndInitSinksOrDie(opt.Sinks, opt.SinkExportDataTimeout)
	//
	//man, err := manager.NewManager(sourceManager, nil, sinkManager,
	//	opt.MetricResolution, manager.DefaultScrapeOffset, manager.DefaultMaxParallelism)
	//if err != nil {
	//	glog.Fatalf("Failed to create main manager: %v", err)
	//}
	//man.Start()
	//waitForStop()
}

func testDiscovery() {
	var sampleFile = `
global:
  discovery_interval: 5m
plugin_configs:
  - type: telegraf/redis
    images:
    - 'redis:*'
    - '*redis*'
    port: 6379
    scheme: "tcp"
    conf: |
      servers = ["${server}"]
      password = "bar"
  - type: telegraf/memcached
    images:
    - 'memcached:*'
    port: 11211
    conf: |
      servers: ${server}
`
	cfg, err := pluginsDiscovery.FromYAML([]byte(sampleFile))
	if err != nil {
		glog.Fatalf("error loading discovery: %q", err)
	}
	//encoder := telegraf.NewEncoder()

	for _, pluginCfg := range cfg.PluginConfigs {
		u, err := url.Parse("?")
		if err != nil {
			glog.Fatalf("error parsing url: %q", err)
		}
		v := url.Values{}
		v.Add("plugins", strings.Replace(pluginCfg.Type, "telegraf/", "", -1))
		v.Add("tg.conf", pluginCfg.Conf)
		u.RawQuery = v.Encode()

		fmt.Println("url", u.String())

		if _, err := pluginsTelegraf.NewFactory().Build(u); err != nil {
			glog.Errorf("error creating telegraf plugin: %q", err)
		}

		//encoding := encoder.Encode("123", discovery.PodType.String(), metav1.ObjectMeta{}, pluginCfg)
		//fmt.Printf("encoding: %s", encoding)
		//u, err := url.Parse(encoding)
		//if err != nil {
		//	glog.Error(err)
		//	return
		//}
		//pluginsTelegraf.NewFactory().Build(u)
	}
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
