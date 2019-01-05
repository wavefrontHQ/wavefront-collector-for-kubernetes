package prometheus

import (
	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

// handles a single prometheus discovery rule
type ruleHandler struct {
	delegate *discoverer
	th       *targetHandler
}

func newRuleHandler(d *discoverer) *ruleHandler {
	return &ruleHandler{
		delegate: d,
		th:       newTargetHandler(d),
	}
}

func (rh *ruleHandler) handle(cfg discovery.PrometheusConfig) error {
	// default to pod
	if cfg.ResourceType == "" {
		cfg.ResourceType = discovery.PodType.String()
	}
	glog.V(4).Infof("rule=%s type=%s labels=%v", cfg.Name, cfg.ResourceType, cfg.Labels)

	// build a new set of targets
	targets := make(map[string]bool)
	switch cfg.ResourceType {
	case discovery.PodType.String():
		rh.findPods(cfg, targets)
	case discovery.ServiceType.String():
		rh.findServices(cfg, targets)
	default:
		glog.Errorf("unknown type: %s for rule: %s", cfg.ResourceType, cfg.Name)
	}

	// delete providers that no longer apply to the rule
	for k := range rh.th.all() {
		if _, exists := targets[k]; !exists {
			rh.th.unregister(k)
		}
	}
	return nil
}

func (rh *ruleHandler) delete() {
	for k := range rh.th.all() {
		rh.th.unregister(k)
	}
}

func (rh *ruleHandler) findPods(cfg discovery.PrometheusConfig, targets map[string]bool) {
	pods, err := rh.delegate.manager.ListPods(cfg.Namespace, cfg.Labels)
	if err != nil {
		glog.Errorf("error listing pods: %v", err)
		return
	}
	glog.V(4).Infof("rule=%s %d pods found", cfg.Name, len(pods))
	for _, pod := range pods {
		name := resourceName(discovery.PodType.String(), pod.ObjectMeta)
		rh.th.discover(pod.Status.PodIP, discovery.PodType.String(), pod.ObjectMeta, cfg)
		targets[name] = true
	}
}

func (rh *ruleHandler) findServices(cfg discovery.PrometheusConfig, targets map[string]bool) {
	services, err := rh.delegate.manager.ListServices(cfg.Namespace, cfg.Labels)
	if err != nil {
		glog.Errorf("error listing services: %v", err)
		return
	}
	glog.V(4).Infof("rule=%s %d services found", cfg.Name, len(services))
	for _, service := range services {
		name := resourceName(discovery.ServiceType.String(), service.ObjectMeta)
		rh.th.discover(service.Spec.ClusterIP, discovery.ServiceType.String(), service.ObjectMeta, cfg)
		targets[name] = true
	}
}
