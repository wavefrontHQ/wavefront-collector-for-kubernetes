package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handles a single prometheus discovery rule
type ruleHandler struct {
	lister discovery.ResourceLister
	th     discovery.TargetHandler
	daemon bool
}

// Gets a new prometheus rule handler
func NewRuleHandler(rl discovery.ResourceLister, ph metrics.ProviderHandler, daemon bool) discovery.RuleHandler {
	return &ruleHandler{
		lister: rl,
		th:     NewTargetHandler(ph, false),
		daemon: daemon,
	}
}

func (rh *ruleHandler) Handle(cfg interface{}) error {
	rule := cfg.(discovery.PrometheusConfig)

	// default to pod
	if rule.ResourceType == "" {
		rule.ResourceType = discovery.PodType.String()
	}
	glog.V(4).Infof("rule=%s type=%s labels=%v", rule.Name, rule.ResourceType, rule.Labels)

	// build a new set of targets
	targets := make(map[string]bool)
	switch rule.ResourceType {
	case discovery.PodType.String():
		rh.findPods(rule, targets)
	case discovery.ServiceType.String():
		rh.findServices(rule, targets)
	case discovery.ApiServerType.String():
		rh.discoverAPIServer(rule, targets)
	default:
		glog.Errorf("unknown type=%s for rule=%s", rule.ResourceType, rule.Name)
	}

	// delete targets that no longer apply to this rule
	rh.th.DeleteMissing(targets)

	return nil
}

func (rh *ruleHandler) Delete() {
	// delete all targets
	rh.th.DeleteMissing(nil)
}

func (rh *ruleHandler) discoverAPIServer(rule discovery.PrometheusConfig, targets map[string]bool) {
	if rh.daemon && !leadership.Leading() {
		glog.V(2).Infof("apiserver discovery disabled. current leader: %s", leadership.Leader())
		return
	}
	if rule.Port == "" {
		rule.Port = "443"
	}
	if rule.Scheme == "" {
		rule.Scheme = "https"
	}
	rh.discover(kubernetes.DefaultAPIService, discovery.ApiServerType.String(),
		metav1.ObjectMeta{Name: "kube-apiserver"}, rule, targets)
}

func (rh *ruleHandler) findPods(rule discovery.PrometheusConfig, targets map[string]bool) {
	if rule.Namespace == "" && len(rule.Labels) == 0 {
		glog.Errorf("error processing rule=%s err=missing namespace and labels", rule.Name)
		return
	}
	pods, err := rh.lister.ListPods(rule.Namespace, rule.Labels)
	if err != nil {
		glog.Errorf("rule=%s error listing pods: %v", rule.Name, err)
		return
	}
	glog.V(4).Infof("rule=%s %d pods found", rule.Name, len(pods))
	for _, pod := range pods {
		rh.discover(pod.Status.PodIP, discovery.PodType.String(), pod.ObjectMeta, rule, targets)
	}
}

func (rh *ruleHandler) findServices(rule discovery.PrometheusConfig, targets map[string]bool) {
	if rh.daemon && !leadership.Leading() {
		glog.V(2).Infof("service discovery disabled. current leader: %s", leadership.Leader())
		return
	}
	if rule.Namespace == "" && len(rule.Labels) == 0 {
		glog.Errorf("error processing rule=%s err=missing namespace and labels", rule.Name)
		return
	}
	services, err := rh.lister.ListServices(rule.Namespace, rule.Labels)
	if err != nil {
		glog.Errorf("rule=%s error listing services: %v", rule.Name, err)
		return
	}
	glog.V(4).Infof("rule=%s %d services found", rule.Name, len(services))
	for _, service := range services {
		rh.discover(service.Spec.ClusterIP, discovery.ServiceType.String(), service.ObjectMeta, rule, targets)
	}
}

func (rh *ruleHandler) discover(ip, kind string, meta metav1.ObjectMeta, rule discovery.PrometheusConfig, targets map[string]bool) {
	name := discovery.ResourceName(kind, meta)
	targets[name] = true
	rh.th.Handle(discovery.Resource{IP: ip, Kind: kind, Meta: meta}, rule)
}
