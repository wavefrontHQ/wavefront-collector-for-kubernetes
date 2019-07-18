package discovery

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"

	"github.com/golang/glog"
)

type ProviderInfo struct {
	Factory metrics.ProviderFactory
	Encoder Encoder
}

type defaultHandler struct {
	info           ProviderInfo
	registry       TargetRegistry
	rh             func(r Resource) bool
	useAnnotations bool

	mtx     sync.RWMutex
	targets map[string]string
}

type HandlerOption func(TargetHandler)

func UseAnnotations(use bool) HandlerOption {
	return func(handler TargetHandler) {
		if h, ok := handler.(*defaultHandler); ok {
			h.useAnnotations = use
		}
	}
}

func SetRegistrationHandler(f func(resource Resource) bool) HandlerOption {
	return func(handler TargetHandler) {
		if h, ok := handler.(*defaultHandler); ok {
			h.rh = f
		}
	}
}

// Gets a new target handler for handling discovered targets
func NewHandler(info ProviderInfo, registry TargetRegistry, setters ...HandlerOption) TargetHandler {
	handler := &defaultHandler{
		info:     info,
		registry: registry,
		targets:  make(map[string]string),
	}
	for _, setter := range setters {
		setter(handler)
	}
	return handler
}

func (d *defaultHandler) Encoding(name string) string {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.targets[name]
}

func (d *defaultHandler) add(name, url string) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.targets[name] = url
}

func (d *defaultHandler) Delete(name string) {
	d.unregister(name)
}

// deletes targets that do not exist in the input map
func (d *defaultHandler) DeleteMissing(input map[string]bool) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	for k := range d.targets {
		if _, exists := input[k]; !exists {
			// delete directly rather than call unregister to prevent recursive locking
			delete(d.targets, k)
			d.deleteProvider(k)
		}
	}
}

func (d *defaultHandler) Handle(resource Resource, rule interface{}) {
	kind := resource.Kind
	ip := resource.IP
	meta := resource.Meta
	glog.Infof("%s: %s namespace: %s", kind, meta.Name, meta.Namespace)

	name := ResourceName(kind, meta)
	currEncoding := d.registry.Encoding(name)

	values := d.info.Encoder.Encode(ip, kind, meta, rule)
	newEncoding := values.Encode()

	// add target if newEncoding is non-empty and has changed
	if len(newEncoding) > 0 && newEncoding != currEncoding {
		glog.V(4).Infof("newEncoding: %s", newEncoding)
		glog.V(4).Infof("currEncoding: %s", currEncoding)

		u, err := url.Parse("?")
		if err != nil {
			glog.Errorf("error parsing url: %q", err)
			return
		}
		u.RawQuery = newEncoding

		provider, err := d.info.Factory.Build(u)
		if err != nil {
			glog.Error(err)
			return
		}
		if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
			i.Configure(u)
		}
		d.register(name, newEncoding, provider)
	}

	// delete target if scrape annotation is false/absent and handler is annotation based
	if newEncoding == "" && currEncoding != "" && d.useAnnotations && d.Encoding(name) != "" {
		if d.rh != nil && d.rh(resource) {
			glog.V(2).Infof("deleting target %s as annotation has changed", name)
			d.unregister(name)
		}
	}
}

func (d *defaultHandler) register(name, url string, provider metrics.MetricsSourceProvider) {
	d.add(name, url)
	sources.Manager().AddProvider(provider)
	d.registry.Register(name, d)
}

func (d *defaultHandler) unregister(name string) {
	d.mtx.Lock()
	delete(d.targets, name)
	d.mtx.Unlock()
	d.deleteProvider(name)
}

func (d *defaultHandler) deleteProvider(name string) {
	if d.registry.Handler(name) != nil {
		providerName := fmt.Sprintf("%s: %s", d.info.Factory.Name(), name)
		sources.Manager().DeleteProvider(providerName)
		d.registry.Unregister(name)
	}
	glog.V(5).Infof("%s deleted", name)
}

func (d *defaultHandler) Count() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return len(d.targets)
}
