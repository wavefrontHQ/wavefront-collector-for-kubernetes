package discovery

import (
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	discoveryEnabled gm.Counter
	rulesCount       gm.Gauge
)

func init() {
	discoveryEnabled = gm.GetOrRegisterCounter("discovery.enabled", gm.DefaultRegistry)
	rulesCount = gm.GetOrRegisterGauge("discovery.rules.count", gm.DefaultRegistry)
}

type discoveryManager struct {
	modTime         time.Time
	daemon          bool
	kubeClient      kubernetes.Interface
	providerHandler metrics.ProviderHandler
	discoverer      discovery.Discoverer
	ruleHandler     discovery.RuleHandler
	serviceListener *serviceHandler
	channel         chan struct{}

	mtx   sync.Mutex
	rules map[string]bool
}

func NewDiscoveryManager(client kubernetes.Interface, cfgFile string, handler metrics.ProviderHandler, daemon bool) {
	mgr := &discoveryManager{
		daemon:          daemon,
		kubeClient:      client,
		providerHandler: handler,
		channel:         make(chan struct{}),
		rules:           make(map[string]bool),
	}

	// load config to init runtime discovery
	var plugins []discovery.PluginConfig
	if cfgFile != "" {
		cfg, err := FromFile(cfgFile)
		if err != nil {
			glog.Fatalf("invalid discovery file: %q", err)
		}
		plugins = cfg.PluginConfigs
	}
	mgr.discoverer = newDiscoverer(handler, plugins)
	mgr.ruleHandler = newRuleHandler(mgr.discoverer, handler, daemon)

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

	// delete rules that were removed/renamed
	rules := make(map[string]bool, len(cfg.PluginConfigs))
	for _, rule := range cfg.PluginConfigs {
		rules[rule.Name] = true
	}
	dm.pruneRules(rules)

	// process current rules
	for _, rule := range cfg.PluginConfigs {
		_, exists := dm.rules[rule.Name]
		if !exists {
			dm.rules[rule.Name] = true
		}
		err := dm.ruleHandler.Handle(rule)
		if err != nil {
			glog.Errorf("error processing rule=%s type=%s err=%v", rule.Name, rule.Type, err)
		}
	}
	rulesCount.Update(int64(len(cfg.PluginConfigs)))
}

func (dm *discoveryManager) pruneRules(newRules map[string]bool) {
	for name := range dm.rules {
		if _, exists := newRules[name]; !exists {
			glog.V(2).Infof("deleting discovery rule %s", name)
			delete(dm.rules, name)
			dm.ruleHandler.Delete(name)
		}
	}
}
