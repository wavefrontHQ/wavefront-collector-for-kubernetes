package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

func TestDiscover(t *testing.T) {
	ph := &util.DummyProviderHandler{}
	th := newTargetHandler(ph)

	pod := discovery.FakePod("pod1", "ns", "123")
	th.discover("123", discovery.PodType.String(), pod.ObjectMeta, discovery.PrometheusConfig{})

	// should not be discovered without annotations
	if len(th.targets) != 0 {
		t.Error("unexpected pod1 registration")
	}

	// add annotations and discover again
	pod.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
	}
	th.discover("123", discovery.PodType.String(), pod.ObjectMeta, discovery.PrometheusConfig{})

	// should be discovered
	if len(th.targets) != 1 {
		t.Error("expected pod1 registration")
	}
}
