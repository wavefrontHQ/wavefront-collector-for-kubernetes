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

type targetHandler struct {
	ph             metrics.DynamicProviderHandler
	useAnnotations bool
	mtx            sync.RWMutex
	targets        map[string]string
}

func newTargetHandler(providerHandler metrics.DynamicProviderHandler) *targetHandler {
	return &targetHandler{
		ph:      providerHandler,
		targets: make(map[string]string),
	}
}

// deletes targets that do not exist in the input map
func (th *targetHandler) deleteMissing(input map[string]bool) {
	th.mtx.Lock()
	defer th.mtx.Unlock()
	for k := range th.targets {
		if _, exists := input[k]; !exists {
			// delete directly rather than call unregister to prevent recursive locking
			delete(th.targets, k)
			th.deleteProvider(k)
		}
	}
}

func (th *targetHandler) get(name string) string {
	th.mtx.RLock()
	defer th.mtx.RUnlock()
	return th.targets[name]
}

func (th *targetHandler) add(name, url string) {
	th.mtx.Lock()
	defer th.mtx.Unlock()
	th.targets[name] = url
}

func (th *targetHandler) discover(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig) {
	glog.V(5).Infof("%s: %s namespace: %s", kind, meta.Name, meta.Namespace)
	name := resourceName(kind, meta)
	cachedURL := registry.registeredURL(name)
	scrapeURL := scrapeURL(ip, kind, meta, rule)

	// add target if scrapeURL is non-empty and has changed
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

	// delete target if scrape annotation is false/absent and handler is annotation based
	if scrapeURL == "" && cachedURL != "" && th.useAnnotations && th.get(name) != "" {
		if param(meta, scrapeAnnotation, "", "false") == "false" {
			glog.V(2).Infof("deleting target %s as scrape is false.", name)
			th.unregister(name)
		}
	}
}

func (th *targetHandler) register(name, url string, provider metrics.MetricsSourceProvider) {
	th.add(name, url)
	th.ph.AddProvider(provider)
	registry.register(name, th)
}

func (th *targetHandler) unregister(name string) {
	th.mtx.Lock()
	delete(th.targets, name)
	th.mtx.Unlock()
	th.deleteProvider(name)
}

func (th *targetHandler) deleteProvider(name string) {
	if registry.registered(name) != nil {
		providerName := fmt.Sprintf("%s: %s", prometheus.ProviderName, name)
		th.ph.DeleteProvider(providerName)
		registry.unregister(name)
	}
	glog.V(5).Infof("%s deleted", name)
}
