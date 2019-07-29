package discovery

import (
	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handles runtime changes to plugin rules
type ruleHandler struct {
	d          *discoverer
	daemon     bool
	rulesCount gm.Gauge
}

// Gets a new rule handler that can handle runtime changes to plugin rules
func newRuleHandler(d discovery.Discoverer, daemon bool) discovery.RuleHandler {
	return &ruleHandler{
		d:          d.(*discoverer),
		daemon:     daemon,
		rulesCount: gm.GetOrRegisterGauge("discovery.rules.count", gm.DefaultRegistry),
	}
}

func (rh *ruleHandler) HandleAll(plugins []discovery.PluginConfig) error {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()

	// delete rules that were removed/renamed
	rules := make(map[string]bool, len(plugins))
	for _, rule := range plugins {
		rules[rule.Name] = true
	}
	for name := range rh.d.delegates {
		if _, exists := rules[name]; !exists {
			log.Infof("deleting discovery rule %s", name)
			rh.internalDelete(name)
		}
	}

	// process current rules
	for _, rule := range plugins {
		err := rh.internalHandle(rule)
		if err != nil {
			log.Errorf("error processing rule=%s type=%s err=%v", rule.Name, rule.Type, err)
		}
	}
	rh.rulesCount.Update(int64(len(plugins)))
	return nil
}

func (rh *ruleHandler) Handle(plugin discovery.PluginConfig) error {
	log.Infof("handling rule=%s type=%s", plugin.Name, plugin.Type)
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	return rh.internalHandle(plugin)
}

func (rh *ruleHandler) DeleteAll() {
	err := rh.HandleAll([]discovery.PluginConfig{})
	if err != nil {
		log.Errorf("error deleting rules: %v", err)
	}
	for _, rh := range rh.d.runtimeHandlers {
		rh.DeleteMissing(map[string]bool{})
	}
}

func (rh *ruleHandler) Delete(name string) {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	rh.internalDelete(name)
}

func (rh *ruleHandler) Count() int {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()
	return len(rh.d.delegates)
}

// this method should only be invoked after acquiring the internal lock
func (rh *ruleHandler) internalHandle(plugin discovery.PluginConfig) error {
	delegate, exists := rh.d.delegates[plugin.Name]
	if !exists {
		var err error
		delegate, err = makeDelegate(plugin)
		if err != nil {
			return err
		}
		rh.d.delegates[plugin.Name] = delegate
	} else {
		// replace the delegate plugin and filter without changing the handler
		filter, err := newResourceFilter(plugin)
		if err != nil {
			return err
		}
		delegate.filter = filter
		delegate.plugin = plugin
	}
	if plugin.Selectors.ResourceType == discovery.ApiServerType.String() {
		rh.discoverAPIServer(plugin, delegate.handler)
	}
	return nil
}

// this method should only be invoked after acquiring the internal lock
func (rh *ruleHandler) internalDelete(name string) {
	// deletes relevant discoverer delegate
	if delegate, exists := rh.d.delegates[name]; exists {
		delegate.handler.DeleteMissing(nil)
		delete(rh.d.delegates, name)
	}
}

func (rh *ruleHandler) discoverAPIServer(plugin discovery.PluginConfig, handler discovery.TargetHandler) {
	if rh.daemon && !leadership.Leading() {
		log.Infof("apiserver discovery disabled. current leader: %s", leadership.Leader())
		return
	}

	if plugin.Port != "" {
		plugin.Port = "443"
	}
	if plugin.Scheme != "https" {
		plugin.Scheme = "https"
	}
	resource := discovery.Resource{
		Kind: discovery.ApiServerType.String(),
		IP:   kubernetes.DefaultAPIService,
		Meta: metav1.ObjectMeta{Name: "kube-apiserver"},
	}
	handler.Handle(resource, plugin)
}
