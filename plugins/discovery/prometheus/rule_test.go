package prometheus

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

func TestAdd(t *testing.T) {
	rh := handler(2)
	rh.Handle(discovery.PrometheusConfig{Name: "test", Namespace: "ns"})
	if rh.th.Count() != 2 {
		t.Errorf("rule add error expected: 2 actual: %d", rh.th.Count())
	}

	rh.lister = discovery.NewFakeResourceLister(4)
	rh.Handle(discovery.PrometheusConfig{Name: "test", Namespace: "ns"})
	if rh.th.Count() != 4 {
		t.Errorf("rule add error expected: 4 actual: %d", rh.th.Count())
	}
	// clear registry
	rh.Delete()
}

func TestDelete(t *testing.T) {
	rh := handler(4)
	rh.Handle(discovery.PrometheusConfig{Name: "test", Namespace: "ns"})
	if rh.th.Count() != 4 {
		t.Errorf("rule delete error expected: 4 actual: %d", rh.th.Count())
	}

	rh.lister = discovery.NewFakeResourceLister(2)
	rh.Handle(discovery.PrometheusConfig{Name: "test", Namespace: "ns"})
	if rh.th.Count() != 2 {
		t.Errorf("rule delete error expected: 2 actual: %d", rh.th.Count())
	}
	// clear registry
	rh.Delete()
}

func handler(count int) *ruleHandler {
	return &ruleHandler{
		lister: discovery.NewFakeResourceLister(count),
		th:     NewTargetHandler(&util.DummyProviderHandler{}, false),
	}
}
