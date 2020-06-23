package discovery

import (
	"reflect"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	discoveryAnnotation = "wavefront.com/discovery-config"
)

type configHandler struct {
	ch       chan struct{}
	informer cache.SharedInformer

	mtx         sync.RWMutex
	cfg         discovery.Config            // main configuration obtained by combining wired and dynamic configuration
	wiredCfg    discovery.Config            // wired configuration
	runtimeCfgs map[string]discovery.Config // dynamic runtime configurations
	changed     bool                        // flag for tracking runtime cfg changes
}

func newConfigHandler(kubeClient kubernetes.Interface, cfg discovery.Config) *configHandler {
	handler := &configHandler{
		cfg:         cfg,
		wiredCfg:    cfg,
		runtimeCfgs: make(map[string]discovery.Config),
	}

	ns := util.GetNamespaceName()
	if ns == "" {
		return handler
	}

	s := kubeClient.CoreV1().ConfigMaps(ns)
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return s.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.Watch(options)
		},
	}

	inf := cache.NewSharedInformer(lw, &v1.ConfigMap{}, 1*time.Hour)
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cfg := obj.(*v1.ConfigMap)
			handler.updated(cfg)
		},
		UpdateFunc: func(_, obj interface{}) {
			cfg := obj.(*v1.ConfigMap)
			handler.updated(cfg)
		},
		DeleteFunc: func(obj interface{}) {
			cfg := obj.(*v1.ConfigMap)
			handler.deleted(cfg)
		},
	})

	handler.informer = inf
	return handler
}

// Config gets the combined discovery configuration and a boolean indicating whether
// the configuration has changed since the last call to this function
func (handler *configHandler) Config() (discovery.Config, bool) {
	handler.mtx.Lock()
	defer handler.mtx.Unlock()

	if !handler.changed {
		return handler.cfg, false
	}
	handler.changed = false

	newCfg := combine(handler.wiredCfg, handler.runtimeCfgs)
	if !reflect.DeepEqual(handler.cfg, newCfg) {
		// update the main combined config
		handler.cfg = newCfg
		return handler.cfg, true
	}
	return handler.cfg, false
}

func (handler *configHandler) updated(cmap *v1.ConfigMap) {
	if !annotated(cmap.GetAnnotations()) {
		// delegate to deleted and return
		log.Infof("no runtime annotation on %s", cmap.Name)
		handler.deleted(cmap)
		return
	}

	loaded, err := load(cmap)
	if err != nil {
		log.Errorf("error loading discovery config: %s error: %v", cmap.Name, err)
		return
	}
	log.Infof("loaded discovery configuration from %s", cmap.Name)

	handler.mtx.Lock()
	defer handler.mtx.Unlock()

	// update the internal map entry
	handler.runtimeCfgs[cmap.Name] = loaded
	handler.changed = true
}

func (handler *configHandler) deleted(cmap *v1.ConfigMap) {
	handler.mtx.Lock()
	defer handler.mtx.Unlock()
	if _, found := handler.runtimeCfgs[cmap.Name]; found {
		log.Infof("deleted discovery configuration from %s", cmap.Name)
		delete(handler.runtimeCfgs, cmap.Name)
		handler.changed = true
	}
}

func annotated(annotations map[string]string) bool {
	if val, ok := annotations[discoveryAnnotation]; ok {
		return val == "true"
	}
	return false
}

func load(cmap *v1.ConfigMap) (discovery.Config, error) {
	cfg := &discovery.Config{}
	for _, data := range cmap.Data {
		loadedCfg, err := discovery.FromYAML([]byte(data))
		if err != nil {
			return *cfg, err
		}
		cfg.PluginConfigs = append(cfg.PluginConfigs, loadedCfg.PluginConfigs...)
	}
	return *cfg, nil
}

func combine(cfg discovery.Config, cfgs map[string]discovery.Config) discovery.Config {
	runCfg := &discovery.Config{
		DiscoveryInterval: cfg.DiscoveryInterval,
		AnnotationPrefix:  cfg.AnnotationPrefix,
		PluginConfigs:     cfg.PluginConfigs,
	}

	// build a sorted slice of map keys for consistent iteration order
	keys := make([]string, len(cfgs))
	i := 0
	for k := range cfgs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	log.Debug("combining discovery configurations")
	for _, key := range keys {
		c := cfgs[key]
		if len(c.PluginConfigs) > 0 {
			runCfg.PluginConfigs = append(runCfg.PluginConfigs, c.PluginConfigs...)
		}
	}
	log.Debugf("total plugin configs: %d", len(runCfg.PluginConfigs))
	return *runCfg
}

func (handler *configHandler) start() bool {
	handler.ch = make(chan struct{})
	go handler.informer.Run(handler.ch)
	return cache.WaitForCacheSync(handler.ch, handler.informer.HasSynced)
}

func (handler *configHandler) stop() {
	if handler.ch != nil {
		close(handler.ch)
	}
}
