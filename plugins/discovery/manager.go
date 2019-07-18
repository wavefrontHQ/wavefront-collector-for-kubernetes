package discovery

import (
	"time"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"k8s.io/client-go/kubernetes"
)

var (
	discoveryEnabled gm.Counter
)

func init() {
	discoveryEnabled = gm.GetOrRegisterCounter("discovery.enabled", gm.DefaultRegistry)
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
}

func NewDiscoveryManager(client kubernetes.Interface, plugins []discovery.PluginConfig, handler metrics.ProviderHandler, daemon bool) *Manager {
	mgr := &Manager{
		daemon:          daemon,
		kubeClient:      client,
		providerHandler: handler,
		stopCh:          make(chan struct{}),
		discoverer:      newDiscoverer(handler, plugins),
	}
	mgr.ruleHandler = newRuleHandler(mgr.discoverer, handler, daemon)
	return mgr
}

func (dm *Manager) Start() {
	glog.Infof("Starting discovery manager")
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
	glog.Infof("Stopping discovery manager")
	discoveryEnabled.Dec(1)

	leadership.Unsubscribe()
	dm.podListener.stop()
	dm.serviceListener.stop()
	close(dm.stopCh)

	dm.discoverer.Stop()
	dm.ruleHandler.DeleteAll()
}

// implements ConfigHandler interface for handling configuration changes
func (dm *Manager) Handle(cfg interface{}) {
	switch cfg.(type) {
	case *discovery.Config:
		glog.Infof("discovery configuration changed")
		d := cfg.(*discovery.Config)
		dm.ruleHandler.HandleAll(d.PluginConfigs)
	case *configuration.Config:
		glog.Infof("discoveryManager: collector configuration changed")
		c := cfg.(*configuration.Config)
		dm.ruleHandler.HandleAll(c.DiscoveryConfigs)
	default:
		glog.Errorf("unknown configuration type: %q", cfg)
	}
}
