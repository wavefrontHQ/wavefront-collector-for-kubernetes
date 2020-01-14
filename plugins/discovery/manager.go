// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	gm "github.com/rcrowley/go-metrics"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	subscriberName = "discovery.manager"
)

var (
	discoveryEnabled gm.Counter
)

func init() {
	discoveryEnabled = gm.GetOrRegisterCounter("discovery.enabled", gm.DefaultRegistry)
}

// RunConfig encapsulates the runtime configuration required for a discovery manager
type RunConfig struct {
	KubeClient      kubernetes.Interface
	DiscoveryConfig discovery.Config
	Handler         metrics.ProviderHandler
	Lister          discovery.ResourceLister
	Daemon          bool
}

// Manager manages the discovery of kubernetes targets based on annotations or configuration rules.
type Manager struct {
	runConfig       RunConfig
	discoverer      discovery.Discoverer
	ruleHandler     discovery.RuleHandler
	podListener     *podHandler
	serviceListener *serviceHandler
	leadershipMgr   *leadership.Manager
	stopCh          chan struct{}
}

// NewDiscoveryManager creates a new instance of a discovery manager based on the given configuration.
func NewDiscoveryManager(cfg RunConfig) *Manager {
	mgr := &Manager{
		runConfig:  cfg,
		stopCh:     make(chan struct{}),
		discoverer: newDiscoverer(cfg.Handler, cfg.DiscoveryConfig),
	}
	mgr.ruleHandler = newRuleHandler(mgr.discoverer, cfg)
	mgr.leadershipMgr = leadership.NewManager(mgr, subscriberName, cfg.KubeClient)
	return mgr
}

func (dm *Manager) Start() {
	log.Infof("Starting discovery manager")
	discoveryEnabled.Inc(1)

	dm.stopCh = make(chan struct{})
	dm.resyncRules()

	// init discovery handlers
	dm.podListener = newPodHandler(dm.runConfig.KubeClient, dm.discoverer)
	dm.serviceListener = newServiceHandler(dm.runConfig.KubeClient, dm.discoverer)
	dm.podListener.start()

	if dm.runConfig.Daemon {
		// in daemon mode, service discovery is performed by only one collector agent in a cluster
		// kick off leader election to determine if this agent should handle it
		dm.leadershipMgr.Start()
	} else {
		dm.serviceListener.start()
	}
}

func (dm *Manager) Stop() {
	log.Infof("Stopping discovery manager")
	discoveryEnabled.Dec(1)

	leadership.Unsubscribe(subscriberName)
	dm.podListener.stop()
	dm.serviceListener.stop()
	close(dm.stopCh)

	dm.discoverer.Stop()
	dm.ruleHandler.DeleteAll()
}

func (dm *Manager) Resume() {
	log.Infof("elected leader: %s starting service discovery", leadership.Leader())
	dm.serviceListener.start()
}

func (dm *Manager) Pause() {
	log.Infof("stopping service discovery. new leader: %s", leadership.Leader())
	dm.serviceListener.stop()
}

// implements ConfigHandler interface for handling configuration changes
func (dm *Manager) Handle(cfg interface{}) {
	switch cfg.(type) {
	case *discovery.Config:
		log.Infof("discovery configuration changed")
		d := cfg.(*discovery.Config)
		dm.ruleHandler.HandleAll(d.PluginConfigs)
	case *configuration.Config:
		log.Infof("discoveryManager: collector configuration changed")
		c := cfg.(*configuration.Config)
		dm.ruleHandler.HandleAll(c.DiscoveryConfig.PluginConfigs)
	default:
		log.Errorf("unknown configuration type: %q", cfg)
	}
}

// resyncRules reloads the discovery rules periodically and stops monitoring resources whose
// lables or namespaces no longer match a configured rule
func (dm *Manager) resyncRules() {
	initial := true
	go wait.Until(func() {
		if initial {
			// wait for listers to index pods and services
			initial = false
			time.Sleep(30 * time.Second)
		}
		log.Info("reloading discovery rules")
		err := dm.ruleHandler.HandleAll(dm.runConfig.DiscoveryConfig.PluginConfigs)
		if err != nil {
			log.Errorf("discovery resync error: %v", err)
		}
	}, dm.runConfig.DiscoveryConfig.DiscoveryInterval, dm.stopCh)
}
