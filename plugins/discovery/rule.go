package discovery

import (
	"fmt"

	"github.com/golang/glog"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handles runtime changes to plugin rules
type ruleHandler struct {
	d      *discoverer
	ph     metrics.ProviderHandler
	daemon bool
}

// Gets a new rule handler that can handle runtime changes to plugin rules
func newRuleHandler(d discovery.Discoverer, ph metrics.ProviderHandler, daemon bool) discovery.RuleHandler {
	return &ruleHandler{
		d:      d.(*discoverer),
		ph:     ph,
		daemon: daemon,
	}
}

func (rh *ruleHandler) Handle(cfg interface{}) error {
	plugin, ok := cfg.(discovery.PluginConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type")
	}
	glog.Infof("handling rule=%s type=%s", plugin.Name, plugin.Type)

	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()

	delegate, exists := rh.d.delegates[plugin.Name]
	if !exists {
		var err error
		delegate, err = makeDelegate(rh.ph, plugin)
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

func (rh *ruleHandler) discoverAPIServer(plugin discovery.PluginConfig, handler discovery.TargetHandler) {
	if rh.daemon && !leadership.Leading() {
		glog.V(2).Infof("apiserver discovery disabled. current leader: %s", leadership.Leader())
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

func (rh *ruleHandler) Delete(name string) {
	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()

	// deletes relevant discoverer delegate
	if delegate, exists := rh.d.delegates[name]; exists {
		delegate.handler.DeleteMissing(nil)
		delete(rh.d.delegates, name)
	}
}
