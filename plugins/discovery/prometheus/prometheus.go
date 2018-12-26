package prometheus

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"

	"k8s.io/api/core/v1"
)

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
	return nil
}

func (d *discoverer) discover(pod *v1.Pod, config discovery.PrometheusConfig, checkAnnotation bool) error {
	glog.V(5).Infof("pod added|updated: %s namespace=%s", pod.Name, pod.Namespace)

	scrapeURL, err := scrapeURL(pod, config, checkAnnotation)
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
		registeredURL := d.manager.Registered(pod.Name)
		urlStr := scrapeURL.Query()["url"][0]
		if urlStr != registeredURL {
			glog.V(4).Infof("scrapeURL: %s\nregisteredURL: %s", urlStr, registeredURL)
			d.manager.RegisterProvider(pod.Name, provider, urlStr)
		}
	}
	return nil
}
