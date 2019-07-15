package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/agent"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	discConfig "github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	kube_config "github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/manager"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/processors"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/summary"

	kubeFlag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"
)

var (
	version     string
	commit      string
	discWatcher util.FileWatcher
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

	glog.Infof(strings.Join(os.Args, " "))
	glog.Infof("wavefront-collector version %v", version)

	preRegister(opt)
	cfg := loadConfigOrDie(opt.ConfigFile)
	cflags := convertOrDie(opt, cfg)
	ag := createAgentOrDie(cflags, cfg)
	registerListeners(ag, opt)
	waitForStop()
}

func preRegister(opt *options.CollectorRunOptions) {
	if opt.Daemon {
		nodeName := os.Getenv(util.NodeNameEnvVar)
		if nodeName == "" {
			glog.Fatalf("missing environment variable %s", util.NodeNameEnvVar)
		}
		err := os.Setenv(util.DaemonModeEnvVar, "true")
		if err != nil {
			glog.Fatalf("error setting environment variable %s", util.DaemonModeEnvVar)
		}
		glog.Infof("%s: %s", util.NodeNameEnvVar, nodeName)
	}
	setMaxProcs(opt)
	registerVersion()
}

func createAgentOrDie(opt *options.CollectorRunOptions, cfg *configuration.Config) *agent.Agent {
	// when invoked from cfg reloads original command flags will be missing
	// always read from the environment variable
	opt.Daemon = os.Getenv(util.DaemonModeEnvVar) != ""

	clusterName := ""
	var plugins []discConfig.PluginConfig

	// if config is missing we will use the flags provided
	if cfg != nil {
		clusterName = resolveClusterName(cfg.ClusterName, opt)
		plugins = cfg.DiscoveryConfigs
	}

	// create source and sink managers
	sourceManager := sources.NewSourceManager(opt.Sources)
	sinkManager := createSinkManagerOrDie(opt.Sinks, opt.SinkExportDataTimeout)

	// create data processors
	kubernetesUrl := getKubernetesAddressOrDie(opt.Sources)
	kubeClient := createKubeClientOrDie(kubernetesUrl)
	podLister := getPodListerOrDie(kubeClient)
	dataProcessors := createDataProcessorsOrDie(kubernetesUrl, clusterName, podLister)

	// create discovery manager
	handler := sourceManager.(metrics.ProviderHandler)
	dm := createDiscoveryManagerOrDie(kubeClient, plugins, handler, opt)

	// create uber manager
	man, err := manager.NewManager(sourceManager, dataProcessors, sinkManager,
		opt.MetricResolution, manager.DefaultScrapeOffset, manager.DefaultMaxParallelism)
	if err != nil {
		glog.Fatalf("Failed to create main manager: %v", err)
	}

	// create and start agent
	ag := agent.NewAgent(man, dm)
	ag.Start()
	return ag
}

func loadConfigOrDie(file string) *configuration.Config {
	glog.Infof("loading config: %s", file)

	if file == "" {
		return nil
	}

	cfg, err := configuration.FromFile(file)
	if err != nil {
		glog.Fatalf("error parsing configuration: %v", err)
		return nil
	}
	fillDefaults(cfg)

	if err := validateCfg(cfg); err != nil {
		glog.Fatalf("invalid configuration file: %v", err)
		return nil
	}
	return cfg
}

// use defaults if no values specified in config file
func fillDefaults(cfg *configuration.Config) {
	if cfg.CollectionInterval == 0 {
		cfg.CollectionInterval = 60 * time.Second
	}
	if cfg.SinkExportDataTimeout == 0 {
		cfg.SinkExportDataTimeout = 20 * time.Second
	}
	if cfg.ClusterName == "" {
		cfg.ClusterName = "k8s-cluster"
	}
}

func convertOrDie(opt *options.CollectorRunOptions, cfg *configuration.Config) *options.CollectorRunOptions {
	// omit flags if config file is provided
	if cfg != nil {
		cflags, err := cfg.Convert()
		if err != nil {
			glog.Fatalf("error converting configuration: %v", err)
		}
		glog.Infof("using configuration file, omitting flags")
		return cflags
	}
	addInternalStatsSource(opt)
	return opt
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

func registerListeners(ag *agent.Agent, opt *options.CollectorRunOptions) {
	handler := &reloader{ag: ag}
	if opt.ConfigFile != "" {
		listener := configuration.NewFileListener(handler)
		watcher := util.NewFileWatcher(opt.ConfigFile, listener, 30*time.Second)
		watcher.Watch()
	}
	if opt.EnableDiscovery && opt.DiscoveryConfigFile != "" && opt.ConfigFile == "" {
		listener := discConfig.NewFileListener(handler)
		discWatcher = util.NewFileWatcher(opt.DiscoveryConfigFile, listener, 30*time.Second)
		discWatcher.Watch()
	}
}

func createDiscoveryManagerOrDie(client *kube_client.Clientset, plugins []discConfig.PluginConfig,
	handler metrics.ProviderHandler, opt *options.CollectorRunOptions) *discovery.Manager {
	if opt.EnableDiscovery {
		// backwards compatibility, discovery config was a separate file
		if len(plugins) == 0 && opt.DiscoveryConfigFile != "" {
			plugins = loadPluginsOrDie(opt.DiscoveryConfigFile)
		}
		return discovery.NewDiscoveryManager(client, plugins, handler, opt.Daemon)
	}
	return nil
}

// backwards compatibility. clusterName used to be specified on the sink.
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

func loadPluginsOrDie(file string) []discConfig.PluginConfig {
	cfg, err := discConfig.FromFile(file)
	if err != nil {
		glog.Fatalf("error loading discovery configuration: %v", err)
	}
	return cfg.PluginConfigs
}

func registerVersion() {
	parts := strings.Split(version, ".")
	friendly := fmt.Sprintf("%s.%s%s", parts[0], parts[1], parts[2])
	f, err := strconv.ParseFloat(friendly, 2)
	if err != nil {
		f = 0.0
	}
	m := gm.GetOrRegisterGaugeFloat64("version", gm.DefaultRegistry)
	m.Update(f)
}

// func createSourceManagerOrDie(src flags.Uris, scrapeTimeout time.Duration) metrics.MetricsSource {
// 	sourceFactory := sources.NewSourceFactory()
// 	sourceList := sourceFactory.BuildAll(src)

// 	for _, source := range sourceList {
// 		glog.Infof("Starting with source %s", source.Name())
// 	}

// 	sourceManager, err := sources.NewSourceManager(sourceList, scrapeTimeout)
// 	if err != nil {
// 		glog.Fatalf("Failed to create source manager: %v", err)
// 	}
// 	return sourceManager
// }

func createSinkManagerOrDie(sinkAddresses flags.Uris, sinkExportDataTimeout time.Duration) metrics.DataSink {
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

func getPodListerOrDie(kubeClient *kube_client.Clientset) v1listers.PodLister {
	podLister, err := util.GetPodLister(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create podLister: %v", err)
	}
	return podLister
}

func createKubeClientOrDie(kubernetesUrl *url.URL) *kube_client.Clientset {
	kubeConfig, err := kube_config.GetKubeClientConfig(kubernetesUrl)
	if err != nil {
		glog.Fatalf("Failed to get client config: %v", err)
	}
	return kube_client.NewForConfigOrDie(kubeConfig)
}

func createDataProcessorsOrDie(kubernetesUrl *url.URL, cluster string, podLister v1listers.PodLister) []metrics.DataProcessor {
	labelCopier, err := util.NewLabelCopier(",", []string{}, []string{})
	if err != nil {
		glog.Fatalf("Failed to initialize label copier: %v", err)
	}

	dataProcessors := []metrics.DataProcessor{
		// Convert cumulative to rate
		processors.NewRateCalculator(metrics.RateMetricsMapping),
	}

	podBasedEnricher, err := processors.NewPodBasedEnricher(podLister, labelCopier)
	if err != nil {
		glog.Fatalf("Failed to create PodBasedEnricher: %v", err)
	}
	dataProcessors = append(dataProcessors, podBasedEnricher)

	namespaceBasedEnricher, err := processors.NewNamespaceBasedEnricher(kubernetesUrl)
	if err != nil {
		glog.Fatalf("Failed to create NamespaceBasedEnricher: %v", err)
	}
	dataProcessors = append(dataProcessors, namespaceBasedEnricher)

	// aggregators
	metricsToAggregate := []string{
		metrics.MetricCpuUsageRate.Name,
		metrics.MetricMemoryUsage.Name,
		metrics.MetricCpuRequest.Name,
		metrics.MetricCpuLimit.Name,
		metrics.MetricMemoryRequest.Name,
		metrics.MetricMemoryLimit.Name,
	}

	metricsToAggregateForNode := []string{
		metrics.MetricCpuRequest.Name,
		metrics.MetricCpuLimit.Name,
		metrics.MetricMemoryRequest.Name,
		metrics.MetricMemoryLimit.Name,
		metrics.MetricEphemeralStorageRequest.Name,
		metrics.MetricEphemeralStorageLimit.Name,
	}

	dataProcessors = append(dataProcessors,
		processors.NewPodAggregator(),
		&processors.NamespaceAggregator{
			MetricsToAggregate: metricsToAggregate,
		},
		&processors.NodeAggregator{
			MetricsToAggregate: metricsToAggregateForNode,
		},
		&processors.ClusterAggregator{
			MetricsToAggregate: metricsToAggregate,
		})

	nodeAutoscalingEnricher, err := processors.NewNodeAutoscalingEnricher(kubernetesUrl, labelCopier)
	if err != nil {
		glog.Fatalf("Failed to create NodeAutoscalingEnricher: %v", err)
	}
	dataProcessors = append(dataProcessors, nodeAutoscalingEnricher)

	// this always needs to be the last processor
	wavefrontCoverter, err := summary.NewPointConverter(kubernetesUrl, cluster)
	if err != nil {
		glog.Fatalf("Failed to create WavefrontPointConverter: %v", err)
	}
	dataProcessors = append(dataProcessors, wavefrontCoverter)

	return dataProcessors
}

// Gets the address of the wavefront sink from the list of sink URIs.
func getWavefrontAddress(args flags.Uris) (*url.URL, error) {
	for _, uri := range args {
		if uri.Key == "wavefront" {
			return &uri.Val, nil
		}
	}
	return nil, fmt.Errorf("no wavefront sink found")
}

// Gets the address of the kubernetes source from the list of source URIs.
// Possible kubernetes sources are: 'kubernetes.summary_api'
func getKubernetesAddressOrDie(args flags.Uris) *url.URL {
	for _, uri := range args {
		if strings.SplitN(uri.Key, ".", 2)[0] == "kubernetes" {
			return &uri.Val
		}
	}
	glog.Fatal("no kubernetes source found")
	return nil
}

func validateCfg(cfg *configuration.Config) error {
	if cfg.CollectionInterval < 5*time.Second {
		return fmt.Errorf("metric resolution should not be less than 5 seconds: %d", cfg.CollectionInterval)
	}
	return nil
}

func setMaxProcs(opt *options.CollectorRunOptions) {
	// Allow as many threads as we have cores unless the user specified a value.
	var numProcs int
	if opt.MaxProcs < 1 {
		numProcs = runtime.NumCPU()
	} else {
		numProcs = opt.MaxProcs
	}
	runtime.GOMAXPROCS(numProcs)

	// Check if the setting was successful.
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != numProcs {
		glog.Warningf("Specified max procs of %d but using %d", numProcs, actualNumProcs)
	}
}

func waitForStop() {
	select {}
}

type reloader struct {
	mtx sync.Mutex
	ag  *agent.Agent
}

// Handles changes to collector or discovery configuration
func (r *reloader) Handle(cfg interface{}) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	switch cfg.(type) {
	case *configuration.Config:
		r.handleCollectorCfg(cfg.(*configuration.Config))
	case *discConfig.Config:
		r.ag.Handle(cfg)
	}
}

func (r *reloader) handleCollectorCfg(cfg *configuration.Config) {
	glog.Infof("collector configuration changed")

	fillDefaults(cfg)

	opt, err := cfg.Convert()
	if err != nil {
		glog.Errorf("configuration error: %v", err)
		return
	}

	// stop the previous agent and start a new agent
	r.ag.Stop()
	r.ag = createAgentOrDie(opt, cfg)
}
