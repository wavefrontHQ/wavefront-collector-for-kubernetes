// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	gm "github.com/rcrowley/go-metrics"

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
	KubeClient             kubernetes.Interface
	DiscoveryConfig        discovery.Config
	Handler                metrics.ProviderHandler
	InternalPluginProvider discovery.PluginProvider
	Lister                 discovery.ResourceLister
	ScrapeCluster          bool
}

// Manager manages the discovery of kubernetes targets based on annotations or configuration rules.
type Manager struct {
	runConfig       RunConfig
	discoverer      discovery.Discoverer
	configListener  *configHandler
	podListener     *podHandler
	serviceListener *serviceHandler
	leadershipMgr   *leadership.Manager
	stopCh          chan struct{}
}

// NewDiscoveryManager creates a new instance of a discovery manager based on the given configuration.
func NewDiscoveryManager(cfg RunConfig) *Manager {
	mgr := &Manager{
		runConfig: cfg,
		stopCh:    make(chan struct{}),
	}
	mgr.leadershipMgr = leadership.NewManager(mgr, subscriberName, cfg.KubeClient)
	return mgr
}

func (dm *Manager) Start() {
	log.Infof("Starting discovery manager")
	discoveryEnabled.Inc(1)

	dm.stopCh = make(chan struct{})

	// init configuration file and discoverer
	cfg := dm.runConfig.DiscoveryConfig
	if cfg.EnableRuntimePlugins {
		log.Info("runtime plugins enabled")
		dm.configListener = newConfigHandler(
			dm.runConfig.KubeClient,
			dm.runConfig.DiscoveryConfig,
			dm.runConfig.InternalPluginProvider,
		)
		if !dm.configListener.start() {
			log.Error("timed out waiting for configmap caches to sync")
		}
		cfg = dm.configListener.Config()
	}
	dm.discoverer = newDiscoverer(dm.runConfig.Handler, cfg, dm.runConfig.Lister)
	dm.startResyncConfig()

	// init discovery handlers
	dm.podListener = newPodHandler(dm.runConfig.KubeClient, dm.discoverer)
	dm.serviceListener = newServiceHandler(dm.runConfig.KubeClient, dm.discoverer)
	dm.podListener.start()

    if dm.runConfig.ScrapeCluster {
        dm.leadershipMgr.Start()
    }
}

func (dm *Manager) Stop() {
	log.Infof("Stopping discovery manager")
	discoveryEnabled.Dec(1)

	leadership.Unsubscribe(subscriberName)
	if dm.configListener != nil {
		dm.configListener.stop()
	}
	dm.podListener.stop()
	dm.serviceListener.stop()
	close(dm.stopCh)

	dm.discoverer.Stop()
	dm.discoverer.DeleteAll()
}

func (dm *Manager) Resume() {
	log.Infof("elected leader: %s starting service discovery", leadership.Leader())
	dm.serviceListener.start()
}

func (dm *Manager) Pause() {
	log.Infof("stopping service discovery. new leader: %s", leadership.Leader())
	dm.serviceListener.stop()
}

// startResyncConfig periodically checks for changes to the discovery config.
// It stops monitoring existing resources and reloads the discovery manager on changes
func (dm *Manager) startResyncConfig() {
	if !dm.runConfig.DiscoveryConfig.EnableRuntimePlugins {
		log.Info("runtime plugins disabled")
		return
	}

	interval := dm.runConfig.DiscoveryConfig.DiscoveryInterval
	log.Infof("discovery config interval: %v", interval)

	go NotifyOfChanges(func() discovery.Config {
		log.Info("checking for runtime plugin changes")
		return dm.configListener.Config()
	}, func() {
		log.Info("found new runtime plugins")
		dm.Stop()
		dm.Start()

	}, interval, dm.stopCh)
}

func NotifyOfChanges(get func() discovery.Config, notify func(), interval time.Duration, stopCh chan struct{}) {
	prevVal := get()
	util.Retry(func() {
		val := get()
		if !reflect.DeepEqual(val, prevVal) {
			notify()
		}
	}, interval, stopCh)
}
