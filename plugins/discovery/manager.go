package discovery

import (
	"sync"
	"time"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
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

type Manager struct {
	modTime         time.Time
	daemon          bool
	kubeClient      kubernetes.Interface
	providerHandler metrics.ProviderHandler
	discoverer      discovery.Discoverer
	ruleHandler     discovery.RuleHandler
	podListener     *podHandler
	serviceListener *serviceHandler
	stopCh          chan struct{}

	mtx   sync.Mutex
	rules map[string]bool
}

func NewDiscoveryManager(client kubernetes.Interface, plugins []discovery.PluginConfig,
	handler metrics.ProviderHandler, daemon bool) *Manager {
	mgr := &Manager{
		daemon:          daemon,
		kubeClient:      client,
		providerHandler: handler,
		stopCh:          make(chan struct{}),
		rules:           make(map[string]bool),
		discoverer:      newDiscoverer(handler, plugins),
	}
	mgr.ruleHandler = newRuleHandler(mgr.discoverer, handler, daemon)
	return mgr
}

func (dm *Manager) Start() {
	discoveryEnabled.Inc(1)

	dm.stopCh = make(chan struct{})

	// init discovery handlers
	dm.podListener = newPodHandler(dm.kubeClient, dm.discoverer)
	dm.serviceListener = newServiceHandler(dm.kubeClient, dm.discoverer)
	dm.podListener.start()

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
					case <-dm.stopCh:
						glog.Infof("stopping service discovery")
						return
					}
				}
			}()
		}
	}
}

func (dm *Manager) Stop() {
	// TODO: support stopping / restarting. Need to start / stop leader election too
}

// implements ConfigHandler interface for handling configuration changes
func (dm *Manager) Handle(cfg interface{}) {
	switch cfg.(type) {
	case *discovery.Config:
		glog.Infof("discovery configuration changed")
		d := cfg.(*discovery.Config)
		dm.load(d.PluginConfigs)
	default:
		glog.Errorf("unknown configuration type: %q", cfg)
	}
}

func (dm *Manager) load(plugins []discovery.PluginConfig) {
	glog.V(2).Info("loading discovery configuration")

	dm.mtx.Lock()
	defer dm.mtx.Unlock()

	// delete rules that were removed/renamed
	rules := make(map[string]bool, len(plugins))
	for _, rule := range plugins {
		rules[rule.Name] = true
	}
	dm.pruneRules(rules)

	// process current rules
	for _, rule := range plugins {
		_, exists := dm.rules[rule.Name]
		if !exists {
			dm.rules[rule.Name] = true
		}
		err := dm.ruleHandler.Handle(rule)
		if err != nil {
			glog.Errorf("error processing rule=%s type=%s err=%v", rule.Name, rule.Type, err)
		}
	}
	rulesCount.Update(int64(len(plugins)))
}

func (dm *Manager) pruneRules(newRules map[string]bool) {
	for name := range dm.rules {
		if _, exists := newRules[name]; !exists {
			glog.V(2).Infof("deleting discovery rule %s", name)
			delete(dm.rules, name)
			dm.ruleHandler.Delete(name)
		}
	}
}
