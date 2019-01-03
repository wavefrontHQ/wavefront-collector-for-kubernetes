package prometheus

import (
	"fmt"
	"net/url"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"

	"github.com/golang/glog"
	"github.com/rcrowley/go-metrics"
	"k8s.io/api/core/v1"
)

var (
	rulesCount metrics.Gauge
)

func init() {
	rulesCount = metrics.GetOrRegisterGauge("discovery.prometheus.rules.count", metrics.DefaultRegistry)
}

type discoverer struct {
	manager discovery.Manager
}

func New(manager discovery.Manager) discovery.Discoverer {
	return &discoverer{
		manager: manager,
	}
}

func (d *discoverer) Discover(pod *v1.Pod) error {
	return d.discover(pod, discovery.PrometheusConfig{}, true)
}

func (d *discoverer) Delete(pod *v1.Pod) {
	glog.V(5).Infof("pod deleted: %s", pod.Name)
	if d.manager.Registered(pod.Name) != "" {
		providerName := fmt.Sprintf("%s: %s", prometheus.ProviderName, pod.Name)
		d.manager.UnregisterProvider(pod.Name, providerName)
	}
}

func (d *discoverer) Process(cfg discovery.Config) error {
	glog.V(2).Info("loading discovery configuration")
	if len(cfg.PromConfigs) == 0 {
		glog.V(2).Info("empty prometheus discovery configs")
		return nil
	}
	for _, promCfg := range cfg.PromConfigs {
		glog.V(4).Info("lookup pods for labels=", promCfg.Labels)
		pods, err := d.manager.ListPods(promCfg.Namespace, promCfg.Labels)
		if err != nil {
			glog.Error(err)
			continue
		}
		glog.V(4).Infof("%d pods found", len(pods))
		for _, pod := range pods {
			d.discover(pod, promCfg, false)
		}
	}
	rulesCount.Update(int64(len(cfg.PromConfigs)))
	return nil
}

func (d *discoverer) discover(pod *v1.Pod, config discovery.PrometheusConfig, checkAnnotation bool) error {
	glog.V(5).Infof("pod: %s added | updated namespace: %s", pod.Name, pod.Namespace)

	registeredURL := d.manager.Registered(pod.Name)
	scrapeURL := scrapeURL(pod, config, checkAnnotation)
	if scrapeURL != "" && scrapeURL != registeredURL {
		glog.V(4).Infof("scrapeURL: %s", scrapeURL)
		glog.V(4).Infof("registeredURL: %s", registeredURL)
		u, err := url.Parse(scrapeURL)
		if err != nil {
			glog.Error(err)
			return err
		}
		provider, err := prometheus.NewPrometheusProvider(u)
		if err != nil {
			glog.Error(err)
			return err
		}
		d.manager.RegisterProvider(pod.Name, provider, scrapeURL)
	}
	return nil
}
