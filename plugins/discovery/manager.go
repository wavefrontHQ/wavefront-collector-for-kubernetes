package discovery

import (
	"os"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
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
	kubeClient      kubernetes.Interface
	resourceLister  discovery.ResourceLister
	providerHandler metrics.DynamicProviderHandler
	discoverer      discovery.Discoverer
	serviceListener *serviceHandler
	channel         chan struct{}
	mtx             sync.Mutex
	rules           map[string]discovery.RuleHandler
}

func NewDiscoveryManager(client kubernetes.Interface, podLister v1listers.PodLister,
	serviceLister v1listers.ServiceLister, cfgFile string, handler metrics.DynamicProviderHandler) {
	mgr := &discoveryManager{
		kubeClient:      client,
		resourceLister:  newResourceLister(podLister, serviceLister),
		providerHandler: handler,
		discoverer:      prometheus.NewDiscoverer(handler),
		channel:         make(chan struct{}),
		rules:           make(map[string]discovery.RuleHandler),
	}
	mgr.Run(cfgFile)
}

func (dm *discoveryManager) Run(cfgFile string) {
	discoveryEnabled.Inc(1)

	// init discovery handlers
	newPodHandler(dm.kubeClient, dm.discoverer)
	dm.serviceListener = newServiceHandler(dm.kubeClient, dm.discoverer)

	// service discovery is performed by only one collector agent in a cluster
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

// reloads the rules now and every discovery interval
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
	glog.V(2).Info("loading discovery rules")
	if len(cfg.PromConfigs) == 0 {
		glog.V(2).Info("found no discovery rules")
		return
	}

	dm.mtx.Lock()
	defer dm.mtx.Unlock()

	// delete rules that were removed/renamed
	rules := make(map[string]bool, len(cfg.PromConfigs))
	for _, rule := range cfg.PromConfigs {
		rules[rule.Name] = true
	}
	for name, handler := range dm.rules {
		if _, exists := rules[name]; !exists {
			glog.V(2).Infof("deleting discovery rule %s", name)
			delete(dm.rules, name)
			handler.Delete()
		}
	}

	// process current rules
	for _, rule := range cfg.PromConfigs {
		handler, exists := dm.rules[rule.Name]
		if !exists {
			handler = prometheus.NewRuleHandler(dm.resourceLister, dm.providerHandler)
			dm.rules[rule.Name] = handler
		}
		err := handler.Handle(rule)
		if err != nil {
			glog.Errorf("error processing rule=%s err=%v", rule.Name, err)
		}
	}
	rulesCount.Update(int64(len(cfg.PromConfigs)))
}
