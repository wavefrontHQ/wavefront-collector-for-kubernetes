package prometheus

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

func TestDiscover(t *testing.T) {
	mgr := discovery.NewFakeManager()
	discoverer := New(mgr)

	pod := discovery.FakePod("pod1", "ns", "123")
	err := discoverer.Discover("123", "pod", pod.ObjectMeta)
	if err != nil {
		t.Error(err)
	}

	// should not be discovered without annotations
	if mgr.Registered("pod1") != "" {
		t.Error("unexpected pod1 registration")
	}

	// add annotations and discover again
	pod.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
	}
	err = discoverer.Discover("123", "pod", pod.ObjectMeta)
	if err != nil {
		t.Error(err)
	}
	// should be discovered
	if mgr.Registered("pod1") == "" {
		t.Error("expected pod1 registration")
	}
}

func TestDelete(t *testing.T) {
	mgr := discovery.NewFakeManager()
	mgr.RegisterProvider("pod1", nil, "pod1")

	if mgr.Registered("pod1") == "" {
		t.Error("expected pod1 registration")
	}

	discoverer := New(mgr)
	pod := discovery.FakePod("pod1", "ns", "123")
	discoverer.Delete("pod", pod.ObjectMeta)
	if mgr.Registered("pod1") != "" {
		t.Error("deleted failed")
	}
}

func TestProcess(t *testing.T) {
	mgr := discovery.NewFakeManager()
	discoverer := New(mgr)
	discoverer.Process(discovery.Config{
		PromConfigs: []discovery.PrometheusConfig{{}},
	})
	if mgr.Registered("pod1") == "" || mgr.Registered("pod2") == "" {
		t.Error("expected pod1 and pod2 registration")
	}
}
