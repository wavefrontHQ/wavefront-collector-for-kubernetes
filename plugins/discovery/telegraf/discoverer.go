package telegraf

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"sync"

	"github.com/golang/glog"
)

type delegate struct {
	filter  *resourceFilter
	handler discovery.TargetHandler
	plugin  discovery.PluginConfig
}

type discoverer struct {
	mtx       sync.RWMutex
	delegates map[string]*delegate
}

// Gets a new plugins discoverer
func NewDiscoverer(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) discovery.Discoverer {
	return &discoverer{
		delegates: makeDelegates(handler, plugins),
	}
}

func makeDelegates(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) map[string]*delegate {
	delegates := make(map[string]*delegate)
	for _, plugin := range plugins {
		delegate, err := makeDelegate(handler, plugin)
		if err != nil {
			glog.Errorf("error parsing plugin: %s error: %v", plugin.Type, err)
			continue
		}
		delegates[plugin.Type] = delegate
	}
	return delegates
}

func makeDelegate(handler metrics.ProviderHandler, plugin discovery.PluginConfig) (*delegate, error) {
	filter, err := newResourceFilter(plugin)
	if err != nil {
		return nil, err
	}
	return &delegate{
		handler: newTargetHandler(handler, plugin.Type),
		filter:  filter,
		plugin:  plugin,
	}, nil
}

func (d *discoverer) Discover(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 || resource.IP == "" {
		// only pod discovery is supported here
		return
	}
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Handle(resource, delegate.plugin)
			break
		}
	}
}

func (d *discoverer) Delete(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 {
		// only pod discovery is supported here
		return
	}
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Delete(discovery.ResourceName(resource.Kind, resource.Meta))
			break
		}
	}
}

// handles runtime changes to plugin rules
type ruleHandler struct {
	d  *discoverer
	ph metrics.ProviderHandler
}

// Gets a new telegraf rule handler that can handle runtime changes to plugin rules
func NewRuleHandler(d discovery.Discoverer, ph metrics.ProviderHandler) discovery.RuleHandler {
	return &ruleHandler{
		d:  d.(*discoverer),
		ph: ph,
	}
}

func (rh *ruleHandler) Handle(cfg interface{}) error {
	plugin, ok := cfg.(discovery.PluginConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type")
	}
	glog.Infof("handling rule=%s images=%s", plugin.Type, plugin.Images)

	rh.d.mtx.Lock()
	defer rh.d.mtx.Unlock()

	delegate, exists := rh.d.delegates[plugin.Type]
	if !exists {
		var err error
		delegate, err = makeDelegate(rh.ph, plugin)
		if err != nil {
			return err
		}
		rh.d.delegates[plugin.Type] = delegate
	} else {
		// replace the delegate plugin and filter without changing the handler
		filter, err := newResourceFilter(plugin)
		if err != nil {
			return err
		}
		delegate.filter = filter
		delegate.plugin = plugin
	}
	return nil
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
