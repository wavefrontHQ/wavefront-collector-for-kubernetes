package prometheus

import (
	"sync"

	"github.com/golang/glog"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	rulesCount metrics.Gauge
)

func init() {
	rulesCount = metrics.GetOrRegisterGauge("discovery.prometheus.rules.count", metrics.DefaultRegistry)
}

type discoverer struct {
	manager        discovery.Manager
	runtimeHandler *targetHandler
	configMutex    sync.Mutex
	rules          map[string]*ruleHandler
	targetMutex    sync.RWMutex
	targets        map[string]*targetHandler
}

func New(manager discovery.Manager) discovery.Discoverer {
	d := &discoverer{
		manager: manager,
		targets: make(map[string]*targetHandler),
		rules:   make(map[string]*ruleHandler),
	}
	d.runtimeHandler = newTargetHandler(d)
	return d
}

func (d *discoverer) Discover(ip, kind string, obj metav1.ObjectMeta) {
	d.runtimeHandler.discover(ip, kind, obj, discovery.PrometheusConfig{})
}

func (d *discoverer) Delete(kind string, obj metav1.ObjectMeta) {
	name := resourceName(kind, obj)
	d.runtimeHandler.unregister(name)
}

func (d *discoverer) Process(cfg discovery.Config) {
	glog.V(2).Info("loading discovery configuration")
	if len(cfg.PromConfigs) == 0 {
		glog.V(2).Info("empty prometheus discovery configs")
	}

	d.configMutex.Lock()
	defer d.configMutex.Unlock()

	// delete rules that were removed or renamed
	rules := make(map[string]bool, len(cfg.PromConfigs))
	for _, promCfg := range cfg.PromConfigs {
		rules[promCfg.Name] = true
	}
	for name, handler := range d.rules {
		if _, exists := rules[name]; !exists {
			delete(d.rules, name)
			handler.delete()
		}
	}

	// then process current set of rules
	for _, promCfg := range cfg.PromConfigs {
		handler, exists := d.rules[promCfg.Name]
		if !exists {
			handler = newRuleHandler(d)
			d.rules[promCfg.Name] = handler
		}
		err := handler.handle(promCfg)
		if err != nil {
			glog.Errorf("error processing rule %s err=%v", promCfg.Name, err)
		}
	}
	rulesCount.Update(int64(len(cfg.PromConfigs)))
}

func (d *discoverer) register(name string, th *targetHandler) {
	d.targetMutex.Lock()
	defer d.targetMutex.Unlock()
	d.targets[name] = th
}

func (d *discoverer) unregister(name string) {
	d.targetMutex.Lock()
	defer d.targetMutex.Unlock()
	delete(d.targets, name)
}

func (d *discoverer) registered(name string) *targetHandler {
	d.targetMutex.RLock()
	defer d.targetMutex.RUnlock()
	return d.targets[name]
}

func (d *discoverer) registeredURL(name string) string {
	handler := d.registered(name)
	if handler != nil {
		return handler.get(name).scrapeURL
	}
	return ""
}
