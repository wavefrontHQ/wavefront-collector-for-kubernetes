package discovery

import (
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"os"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/telegraf"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
)

var (
	discoveryEnabled gm.Counter
	promRulesCount   gm.Gauge
	pluginRulesCount gm.Gauge
)

func init() {
	discoveryEnabled = gm.GetOrRegisterCounter("discovery.enabled", gm.DefaultRegistry)

	name := reporting.EncodeKey("discovery.rules.count", map[string]string{"type": "prometheus"})
	promRulesCount = gm.GetOrRegisterGauge(name, gm.DefaultRegistry)

	name = reporting.EncodeKey("discovery.rules.count", map[string]string{"type": "plugins"})
	pluginRulesCount = gm.GetOrRegisterGauge(name, gm.DefaultRegistry)
}

type discoveryManager struct {
	modTime             time.Time
	daemon              bool
	kubeClient          kubernetes.Interface
	resourceLister      discovery.ResourceLister
	providerHandler     metrics.ProviderHandler
	discoverer          discovery.Discoverer
	telegrafRuleHandler discovery.RuleHandler
	serviceListener     *serviceHandler
	channel             chan struct{}

	mtx         sync.Mutex
	promRules   map[string]discovery.RuleHandler
	pluginRules map[string]discovery.RuleHandler
}

func NewDiscoveryManager(client kubernetes.Interface, podLister v1listers.PodLister,
	serviceLister v1listers.ServiceLister, cfgFile string, handler metrics.ProviderHandler, daemon bool) {
	mgr := &discoveryManager{
		daemon:          daemon,
		kubeClient:      client,
		resourceLister:  newResourceLister(podLister, serviceLister),
		providerHandler: handler,
		channel:         make(chan struct{}),
		promRules:       make(map[string]discovery.RuleHandler),
		pluginRules:     make(map[string]discovery.RuleHandler),
	}

	// load config here to init runtime discovery based on container images
	// rule based discovery is handled within the Run() flow
	var plugins []discovery.PluginConfig
	if cfgFile != "" {
		cfg, err := FromFile(cfgFile)
		if err != nil {
			glog.Fatalf("invalid discovery file: %q", err)
		}
		plugins = cfg.PluginConfigs
	}

	prometheusDiscoverer := discovery.NewDiscoverer(prometheus.NewTargetHandler(handler, true))
	telegrafDiscoverer := telegraf.NewDiscoverer(handler, plugins)
	mgr.discoverer = newDiscoverer(prometheusDiscoverer, telegrafDiscoverer)
	mgr.telegrafRuleHandler = telegraf.NewRuleHandler(telegrafDiscoverer, handler)

	// Run the manager
	mgr.Run(cfgFile)
}

func (dm *discoveryManager) Run(cfgFile string) {
	discoveryEnabled.Inc(1)

	// init discovery handlers
	newPodHandler(dm.kubeClient, dm.discoverer)
	dm.serviceListener = newServiceHandler(dm.kubeClient, dm.discoverer)

	if !dm.daemon {
		dm.serviceListener.start()
	} else {
		// in daemon mode, service discovery is performed by only one collector agent in a cluster
		// kick off leader election to determine if this agent should handle it
		ch, err := leadership.Subscribe(dm.kubeClient.CoreV1())
		if err != nil {
			glog.Errorf("discovery: leader election error: %q", err)
		} else {
			go func() {
				for {
					select {
					case isLeader := <-ch:
						if isLeader {
							glog.V(2).Infof("elected leader: %s starting service discovery", leadership.Leader())
							dm.serviceListener.start()
						} else {
							glog.V(2).Infof("stopping service discovery. new leader: %s", leadership.Leader())
							dm.serviceListener.stop()
						}
					}
				}
			}()
		}
	}

	if cfgFile != "" {
		dm.loadWatch(cfgFile)
	}
}

// loads the config into memory and watches for changes once a minute
func (dm *discoveryManager) loadWatch(cfgFile string) {
	initial := true
	go wait.Until(func() {
		if initial {
			// wait for listers to index pods and services
			initial = false
			time.Sleep(30 * time.Second)
		}
		fileInfo, err := os.Stat(cfgFile)
		if err != nil {
			glog.Errorf("error retrieving discovery config file stats: %v", err)
			return
		}

		if fileInfo.ModTime().After(dm.modTime) {
			dm.modTime = fileInfo.ModTime()
			cfg, err := FromFile(cfgFile)
			if err != nil {
				glog.Errorf("error loading discovery config: %v", err)
			} else {
				close(dm.channel)
				dm.channel = make(chan struct{})
				dm.reload(*cfg)
			}
		}
	}, 1*time.Minute, wait.NeverStop)
}

// reloads the promRules now and every discovery interval
func (dm *discoveryManager) reload(cfg discovery.Config) {
	syncInterval := 10 * time.Minute
	if cfg.Global.DiscoveryInterval != 0 {
		syncInterval = cfg.Global.DiscoveryInterval
	}
	go wait.Until(func() {
		dm.load(cfg)
	}, syncInterval, dm.channel)
	glog.V(5).Info("discovery reloading terminated")
}

func (dm *discoveryManager) load(cfg discovery.Config) {
	glog.V(2).Info("loading discovery configuration")

	dm.mtx.Lock()
	defer dm.mtx.Unlock()

	// delete prometheus rules that were removed/renamed
	rules := make(map[string]bool, len(cfg.PromConfigs))
	for _, rule := range cfg.PromConfigs {
		rules[rule.Name] = true
	}
	pruneRules(dm.promRules, rules)

	// process current prometheus rules
	for _, rule := range cfg.PromConfigs {
		handler, exists := dm.promRules[rule.Name]
		if !exists {
			handler = prometheus.NewRuleHandler(dm.resourceLister, dm.providerHandler, dm.daemon)
			dm.promRules[rule.Name] = handler
		}
		err := handler.Handle(rule)
		if err != nil {
			glog.Errorf("error processing rule=%s err=%v", rule.Name, err)
		}
	}
	promRulesCount.Update(int64(len(cfg.PromConfigs)))

	// delete plugin rules that were removed/renamed
	rules = make(map[string]bool, len(cfg.PluginConfigs))
	for _, rule := range cfg.PluginConfigs {
		rules[rule.Type] = true
	}
	pruneRules(dm.pluginRules, rules)

	// process current plugin rules
	for _, rule := range cfg.PluginConfigs {
		_, exists := dm.pluginRules[rule.Type]
		if !exists {
			dm.pluginRules[rule.Type] = dm.telegrafRuleHandler
		}
		err := dm.telegrafRuleHandler.Handle(rule)
		if err != nil {
			glog.Errorf("error processing rule=%s err=%v", rule.Type, err)
		}
	}
	pluginRulesCount.Update(int64(len(cfg.PluginConfigs)))
}

func pruneRules(rules map[string]discovery.RuleHandler, newRules map[string]bool) {
	for name, handler := range rules {
		if _, exists := newRules[name]; !exists {
			glog.V(2).Infof("deleting discovery rule %s", name)
			delete(rules, name)
			handler.Delete(name)
		}
	}
}
