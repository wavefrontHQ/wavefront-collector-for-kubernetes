package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
	"github.com/spf13/pflag"

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

	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
)

var (
	version string
	commit  string
)

func main() {
	opt := options.NewCollectorRunOptions()
	opt.AddFlags(pflag.CommandLine)

	flag.InitFlags()

	if opt.Version {
		fmt.Println(fmt.Sprintf("version: %s\ncommit: %s", version, commit))
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	registerVersion()

	labelCopier, err := util.NewLabelCopier(opt.LabelSeparator, opt.StoredLabels, opt.IgnoredLabels)
	if err != nil {
		glog.Fatalf("Failed to initialize label copier: %v", err)
	}

	setMaxProcs(opt)
	glog.Infof(strings.Join(os.Args, " "))
	glog.Infof("wavefront-collector version %v", version)
	if err := validateFlags(opt); err != nil {
		glog.Fatal(err)
	}

	kubernetesUrl, err := getKubernetesAddress(opt.Sources)
	if err != nil {
		glog.Fatalf("Failed to get kubernetes address: %v", err)
	}
	sourceManager := createSourceManagerOrDie(opt.Sources, opt.InternalStatsPrefix)
	sinkManager := createAndInitSinksOrDie(opt.Sinks, opt.SinkExportDataTimeout)

	kubeClient := createKubeClientOrDie(kubernetesUrl)
	podLister := getPodListerOrDie(kubeClient)
	dataProcessors := createDataProcessorsOrDie(kubernetesUrl, podLister, labelCopier)

	if opt.EnableDiscovery {
		handler := sourceManager.(metrics.DynamicProviderHandler)
		serviceLister := getServiceListerOrDie(kubeClient)
		createDiscoveryManagerOrDie(kubeClient, podLister, serviceLister, opt.DiscoveryConfigFile, handler)
	}

	man, err := manager.NewManager(sourceManager, dataProcessors, sinkManager,
		opt.MetricResolution, manager.DefaultScrapeOffset, manager.DefaultMaxParallelism)
	if err != nil {
		glog.Fatalf("Failed to create main manager: %v", err)
	}
	man.Start()
	waitForStop()
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

func createSourceManagerOrDie(src flags.Uris, statsPrefix string) metrics.MetricsSource {
	sourceFactory := sources.NewSourceFactory()
	sourceList := sourceFactory.BuildAll(src, statsPrefix)

	for _, source := range sourceList {
		glog.Infof("Starting with source %s", source.Name())
	}

	sourceManager, err := sources.NewSourceManager(sourceList, sources.DefaultMetricsScrapeTimeout)
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

func createDiscoveryManagerOrDie(client *kube_client.Clientset, podLister v1listers.PodLister,
	serviceLister v1listers.ServiceLister, cfgFile string, handler metrics.DynamicProviderHandler) {
	discovery.NewDiscoveryManager(client, podLister, serviceLister, cfgFile, handler)
}

func getPodListerOrDie(kubeClient *kube_client.Clientset) v1listers.PodLister {
	podLister, err := util.GetPodLister(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create podLister: %v", err)
	}
	return podLister
}

func getServiceListerOrDie(kubeClient *kube_client.Clientset) v1listers.ServiceLister {
	serviceLister, err := util.GetServiceLister(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create serviceLister: %v", err)
	}
	return serviceLister
}

func createKubeClientOrDie(kubernetesUrl *url.URL) *kube_client.Clientset {
	kubeConfig, err := kube_config.GetKubeClientConfig(kubernetesUrl)
	if err != nil {
		glog.Fatalf("Failed to get client config: %v", err)
	}
	return kube_client.NewForConfigOrDie(kubeConfig)
}

func createDataProcessorsOrDie(kubernetesUrl *url.URL, podLister v1listers.PodLister, labelCopier *util.LabelCopier) []metrics.DataProcessor {
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
	return dataProcessors
}

// Gets the address of the kubernetes source from the list of source URIs.
// Possible kubernetes sources are: 'kubernetes' and 'kubernetes.summary_api'
func getKubernetesAddress(args flags.Uris) (*url.URL, error) {
	for _, uri := range args {
		if strings.SplitN(uri.Key, ".", 2)[0] == "kubernetes" {
			return &uri.Val, nil
		}
	}
	return nil, fmt.Errorf("no kubernetes source found")
}

func validateFlags(opt *options.CollectorRunOptions) error {
	if opt.MetricResolution < 5*time.Second {
		return fmt.Errorf("metric resolution should not be less than 5 seconds - %d", opt.MetricResolution)
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
