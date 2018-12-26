package discovery

import (
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type discoveryManager struct {
	kubeClient      kubernetes.Interface
	cfgModTime      time.Time
	podLister       v1listers.PodLister
	providerHandler metrics.DynamicProviderHandler
	discoverer      discovery.Discoverer
	done            chan struct{}
	channel         chan struct{}
	mtx             sync.RWMutex
	registeredPods  map[string]string
}

func NewDiscoveryManager(client kubernetes.Interface, podLister v1listers.PodLister, cfgFile string,
	handler metrics.DynamicProviderHandler) {
	mgr := &discoveryManager{
		kubeClient:      client,
		podLister:       podLister,
		providerHandler: handler,
		registeredPods:  make(map[string]string),
		done:            make(chan struct{}),
		channel:         make(chan struct{}),
	}
	mgr.Run(cfgFile)
}

func (dm *discoveryManager) Run(cfgFile string) {
	dm.discoverer = prometheus.New(dm)
	p := dm.kubeClient.CoreV1().Pods(apiv1.NamespaceAll)
	plw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.Watch(options)
		},
	}
	podInformer := cache.NewSharedInformer(plw, &apiv1.Pod{}, 110*time.Minute)
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*apiv1.Pod)
			dm.discoverer.Discover(pod)
		},
		UpdateFunc: func(_, obj interface{}) {
			pod := obj.(*apiv1.Pod)
			dm.discoverer.Discover(pod)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*apiv1.Pod)
			dm.discoverer.Delete(pod)
		},
	})
	go podInformer.Run(dm.done)

	if cfgFile != "" {
		dm.load(cfgFile)
	}
}

// loads the cfgFile and checks for changes once a minute
func (dm *discoveryManager) load(cfgFile string) {
	initial := true
	go wait.Until(func() {
		if initial {
			// wait for podLister to index pods
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

func (dm *discoveryManager) RegisterProvider(podName string, provider metrics.MetricsSourceProvider, obj string) {
	dm.providerHandler.AddProvider(provider)
	dm.registerPod(podName, obj)
}

func (dm *discoveryManager) UnregisterProvider(podName, providerName string) {
	glog.V(2).Infof("deleting provider: %s", providerName)
	dm.providerHandler.DeleteProvider(providerName)
	dm.unregisterPod(podName)
}

func (dm *discoveryManager) registerPod(name string, obj string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	dm.registeredPods[name] = obj
}

func (dm *discoveryManager) unregisterPod(name string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	delete(dm.registeredPods, name)
}

func (dm *discoveryManager) Registered(name string) string {
	dm.mtx.RLock()
	defer dm.mtx.RUnlock()
	return dm.registeredPods[name]
}

func (dm *discoveryManager) ListPods(ns string, l map[string]string) ([]*apiv1.Pod, error) {
	if ns == "" {
		return dm.podLister.List(labels.SelectorFromSet(l))
	}
	nsLister := dm.podLister.Pods(ns)
	return nsLister.List(labels.SelectorFromSet(l))
}
