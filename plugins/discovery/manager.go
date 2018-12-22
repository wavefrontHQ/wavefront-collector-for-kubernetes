package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"os"
	"sync"
	"time"
)

type discoveryManager struct {
	kubeClient      kubernetes.Interface
	cfgModTime      time.Time
	podLister       v1listers.PodLister
	sourceFactory   *sources.SourceFactory
	providerHandler metrics.DynamicProviderHandler
	done            chan struct{}
	channel         chan struct{}
	mtx             sync.RWMutex
	registeredPods  map[string]bool
}

func NewDiscoveryManager(client kubernetes.Interface, podLister v1listers.PodLister, cfgFile string,
	handler metrics.DynamicProviderHandler) {

	mgr := &discoveryManager{
		kubeClient:      client,
		podLister:       podLister,
		sourceFactory:   sources.NewSourceFactory(),
		providerHandler: handler,
		registeredPods:  make(map[string]bool),
		done:            make(chan struct{}),
		channel:         make(chan struct{}),
	}
	// TODO: consider calling this from main?
	mgr.Run()
	mgr.loadConfig(cfgFile)
}

// loads the cfgFile and checks for changes once a minute
func (dm *discoveryManager) loadConfig(cfgFile string) {
	go wait.Until(func() {
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
				dm.processConfig(*cfg)
			}
		}
	}, 1*time.Minute, wait.NeverStop)
}

// processes the discovery configuration rules
func (dm *discoveryManager) processConfig(cfg Config) {
	syncInterval := 10 * time.Minute
	if cfg.Global.DiscoveryInterval != 0 {
		syncInterval = cfg.Global.DiscoveryInterval
	}
	go wait.Until(func() {
		dm.processPromConfigs(cfg.PromConfigs)
	}, syncInterval, dm.channel)
	glog.V(8).Info("ended discovery config processing")
}

func (dm *discoveryManager) Run() {
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
			dm.add(pod, PrometheusConfig{}, true)
		},
		UpdateFunc: func(_, obj interface{}) {
			pod := obj.(*apiv1.Pod)
			dm.add(pod, PrometheusConfig{}, true)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*apiv1.Pod)
			dm.delete(pod)
		},
	})
	go podInformer.Run(dm.done)
}

func (dm *discoveryManager) add(pod *apiv1.Pod, config PrometheusConfig, checkScrapeAnnotation bool) error {
	glog.V(5).Infof("pod added/updated: %s namespace=%s", pod.Name, pod.Namespace)

	if dm.registered(pod.Name) {
		glog.Infof("pod already registered %s", pod.Name)
		return fmt.Errorf("pod already registered %s", pod.Name)
	}

	scrapeURL, err := ScrapeURL(pod, config, checkScrapeAnnotation)
	if err != nil {
		glog.Error(err)
		return err
	}
	if scrapeURL != nil {
		provider, err := prometheus.NewPrometheusProvider(scrapeURL)
		if err != nil {
			glog.Error(err)
			return err
		}
		dm.providerHandler.AddProvider(provider)
		dm.registerPod(pod.Name)
	}
	return nil
}

func (dm *discoveryManager) delete(pod *apiv1.Pod) {
	glog.V(5).Infof("pod deleted: ", pod.Name)
	if dm.registered(pod.Name) {
		name := fmt.Sprintf("%s: %s", prometheus.ProviderName, pod.Name)
		glog.V(2).Infof("deleting provider: %s", name)
		dm.providerHandler.DeleteProvider(name)
		dm.unregisterPod(pod.Name)
	}
}

func (dm *discoveryManager) processPromConfigs(promCfgs []PrometheusConfig) {
	//TODO: move into separate prometheus package once interface is fleshed out

	if len(promCfgs) == 0 {
		glog.V(2).Infof("empty prometheus discovery configs")
		return
	}

	//TODO: need to remove pods that may no longer need to be monitored
	// say if labels are updated and registered pods no longer match the criteria
	for _, promCfg := range promCfgs {
		glog.V(5).Info("discovering pods with labels ", promCfg.Labels)
		pods, err := dm.listPods(promCfg)
		if err != nil {
			glog.Error(err)
			continue
		}
		glog.V(5).Infof("%d pods discovered", len(pods))

		for _, pod := range pods {
			dm.add(pod, promCfg, false)
		}
	}
}

func (dm *discoveryManager) registerPod(name string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	dm.registeredPods[name] = true
}

func (dm *discoveryManager) unregisterPod(name string) {
	dm.mtx.Lock()
	defer dm.mtx.Unlock()
	delete(dm.registeredPods, name)
}

func (dm *discoveryManager) registered(name string) bool {
	dm.mtx.RLock()
	defer dm.mtx.RUnlock()
	_, ok := dm.registeredPods[name]
	return ok
}

func (dm *discoveryManager) listPods(cfg PrometheusConfig) ([]*apiv1.Pod, error) {
	if cfg.Namespace == "" {
		return dm.podLister.List(labels.SelectorFromSet(cfg.Labels))
	}
	nsLister := dm.podLister.Pods(cfg.Namespace)
	return nsLister.List(labels.SelectorFromSet(cfg.Labels))
}
