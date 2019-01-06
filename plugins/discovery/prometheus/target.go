package prometheus

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type target struct {
	name      string
	scrapeURL string
}

func newTarget(name, u string) target {
	return target{
		name:      name,
		scrapeURL: u,
	}
}

type targetHandler struct {
	ph      metrics.DynamicProviderHandler
	mtx     sync.RWMutex
	targets map[string]string
}

func newTargetHandler(providerHandler metrics.DynamicProviderHandler) *targetHandler {
	return &targetHandler{
		ph:      providerHandler,
		targets: make(map[string]string),
	}
}

func (th *targetHandler) all() map[string]string {
	th.mtx.RLock()
	defer th.mtx.RUnlock()
	return th.targets
}

func (th *targetHandler) get(name string) target {
	th.mtx.RLock()
	val := th.targets[name]
	th.mtx.RUnlock()

	if val != "" {
		return newTarget(name, val)
	}
	return target{}
}

func (th *targetHandler) add(name, url string) {
	th.mtx.Lock()
	defer th.mtx.Unlock()
	th.targets[name] = url
}

func (th *targetHandler) delete(name string) {
	th.mtx.Lock()
	defer th.mtx.Unlock()
	delete(th.targets, name)
}

func (th *targetHandler) discover(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig) {
	glog.V(5).Infof("%s: %s namespace: %s", kind, meta.Name, meta.Namespace)
	name := resourceName(kind, meta)
	cachedURL := registry.registeredURL(name)
	scrapeURL := scrapeURL(ip, kind, meta, rule)
	if scrapeURL != "" && scrapeURL != cachedURL {
		glog.V(4).Infof("scrapeURL: %s", scrapeURL)
		glog.V(4).Infof("cachedURL: %s", cachedURL)
		u, err := url.Parse(scrapeURL)
		if err != nil {
			glog.Error(err)
			return
		}
		provider, err := prometheus.NewPrometheusProvider(u)
		if err != nil {
			glog.Error(err)
			return
		}
		th.register(name, scrapeURL, provider)
	}
}

func (th *targetHandler) register(name, url string, provider metrics.MetricsSourceProvider) {
	th.add(name, url)
	th.ph.AddProvider(provider)
	registry.register(name, th)
}

func (th *targetHandler) unregister(name string) {
	th.delete(name)
	if registry.registered(name) != nil {
		providerName := fmt.Sprintf("%s: %s", prometheus.ProviderName, name)
		th.ph.DeleteProvider(providerName)
		registry.unregister(name)
	}
	glog.V(5).Infof("%s deleted", name)
}
