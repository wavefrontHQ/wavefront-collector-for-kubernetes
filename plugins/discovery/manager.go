package discovery

import (
	"os"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
)

var (
	objCount         gm.Gauge
	discoveryEnabled gm.Counter
)

func init() {
	objCount = gm.GetOrRegisterGauge("discovery.objects.registered", gm.DefaultRegistry)
	discoveryEnabled = gm.GetOrRegisterCounter("discovery.enabled", gm.DefaultRegistry)
}

type discoveryManager struct {
	kubeClient      kubernetes.Interface
	cfgModTime      time.Time
	podLister       v1listers.PodLister
	serviceLister   v1listers.ServiceLister
	providerHandler metrics.DynamicProviderHandler
	discoverer      discovery.Discoverer
	done            chan struct{}
	channel         chan struct{}
	mtx             sync.RWMutex
	registeredObjs  map[string]string
}

func NewDiscoveryManager(client kubernetes.Interface, podLister v1listers.PodLister,
	serviceLister v1listers.ServiceLister, cfgFile string, handler metrics.DynamicProviderHandler) {
	mgr := &discoveryManager{
		kubeClient:      client,
		podLister:       podLister,
		serviceLister:   serviceLister,
		providerHandler: handler,
		registeredObjs:  make(map[string]string),
		channel:         make(chan struct{}),
	}
	mgr.Run(cfgFile)
}

func (dm *discoveryManager) Run(cfgFile string) {
	discoveryEnabled.Inc(1)
	dm.discoverer = prometheus.New(dm)

	// init discovery handlers
	NewPodHandler(dm.kubeClient, dm.discoverer)
	NewServiceHandler(dm.kubeClient, dm.discoverer)

	// update the internal object counter once a minute
	go wait.Forever(func() {
		dm.mtx.RLock()
		defer dm.mtx.RUnlock()
		objCount.Update(int64(len(dm.registeredObjs)))
	}, 1*time.Minute)

	if cfgFile != "" {
		dm.load(cfgFile)
	}
}

// loads the cfgFile and checks for changes once a minute
func (dm *discoveryManager) load(cfgFile string) {
	initial := true
	go wait.Until(func() {
		if initial {
			// wait for listers to index pods and services
			initial = false
			time.Sleep(30 * time.Second)
		}
		fileInfo, err := os.Stat(cfgFile)
		if err != nil {
			glog.Fatalf("unable to get discovery config file stats: %v", err)
		}

		if fileInfo.ModTime().After(dm.cfgModTime) {
			dm.cfgModTime = fileInfo.ModTime()
			cfg, err := FromFile(cfgFile)
			if err != nil {
				glog.Errorf("unable to load discovery config: %v", err)
			} else {
				close(dm.channel)
				dm.channel = make(chan struct{})
				dm.process(*cfg)
			}
		}
	}, 1*time.Minute, wait.NeverStop)
}

// processes the discovery configuration rules
func (dm *discoveryManager) process(cfg discovery.Config) {
	syncInterval := 10 * time.Minute
	if cfg.Global.DiscoveryInterval != 0 {
		syncInterval = cfg.Global.DiscoveryInterval
	}
	go wait.Until(func() {
		dm.discoverer.Process(cfg)
	}, syncInterval, dm.channel)
	glog.V(8).Info("ended discovery config processing")
}

func (dm *discoveryManager) RegisterProvider(objName string, provider metrics.MetricsSourceProvider, obj string) {
	dm.providerHandler.AddProvider(provider)
	dm.register(objName, obj)
}

func (dm *discoveryManager) UnregisterProvider(objName, providerName string) {
	glog.V(2).Infof("deleting provider: %s", providerName)
	dm.providerHandler.DeleteProvider(providerName)
	dm.unregister(objName)
}

func (dm *discoveryManager) register(name string, obj string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	dm.registeredObjs[name] = obj
}

func (dm *discoveryManager) unregister(name string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	delete(dm.registeredObjs, name)
}

func (dm *discoveryManager) Registered(name string) string {
	dm.mtx.RLock()
	defer dm.mtx.RUnlock()
	return dm.registeredObjs[name]
}

func (dm *discoveryManager) ListPods(ns string, l map[string]string) ([]*apiv1.Pod, error) {
	if ns == "" {
		return dm.podLister.List(labels.SelectorFromSet(l))
	}
	nsLister := dm.podLister.Pods(ns)
	return nsLister.List(labels.SelectorFromSet(l))
}

func (dm *discoveryManager) ListServices(ns string, l map[string]string) ([]*apiv1.Service, error) {
	if ns == "" {
		return dm.serviceLister.List(labels.SelectorFromSet(l))
	}
	nsLister := dm.serviceLister.Services(ns)
	return nsLister.List(labels.SelectorFromSet(l))
}
