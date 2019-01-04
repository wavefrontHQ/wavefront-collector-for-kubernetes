package prometheus

import (
	"testing"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

func TestDiscover(t *testing.T) {
	mgr := discovery.NewFakeManager()
	discoverer := New(mgr)

	pod := discovery.FakePod("pod1", "ns", "123")
	err := discoverer.Discover("123", discovery.PodType.String(), pod.ObjectMeta)
	if err != nil {
		t.Error(err)
	}

	resName := resourceName(discovery.PodType.String(), pod.ObjectMeta)

	// should not be discovered without annotations
	if mgr.Registered(resName) != "" {
		t.Error("unexpected pod1 registration")
	}

	// add annotations and discover again
	pod.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
	}
	err = discoverer.Discover("123", discovery.PodType.String(), pod.ObjectMeta)
	if err != nil {
		t.Error(err)
	}
	// should be discovered
	if mgr.Registered(resName) == "" {
		t.Error("expected pod1 registration")
	}
}

func TestDelete(t *testing.T) {
	pod := discovery.FakePod("pod1", "ns", "123")
	resName := resourceName(discovery.PodType.String(), pod.ObjectMeta)

	mgr := discovery.NewFakeManager()
	mgr.RegisterProvider(resName, nil, "pod1")

	if mgr.Registered(resName) == "" {
		t.Error("expected pod1 registration")
	}

	discoverer := New(mgr)
	discoverer.Delete(discovery.PodType.String(), pod.ObjectMeta)
	if mgr.Registered(resName) != "" {
		t.Error("deleted failed")
	}
}

func TestProcess(t *testing.T) {
	mgr := discovery.NewFakeManager()
	discoverer := New(mgr)
	discoverer.Process(discovery.Config{
		PromConfigs: []discovery.PrometheusConfig{{}},
	})
	pod1 := discovery.FakePod("pod1", "ns", "1234")
	pod2 := discovery.FakePod("pod2", "ns", "1235")
	pod1Name := resourceName(discovery.PodType.String(), pod1.ObjectMeta)
	pod2Name := resourceName(discovery.PodType.String(), pod2.ObjectMeta)
	if mgr.Registered(pod1Name) == "" || mgr.Registered(pod2Name) == "" {
		t.Error("expected pod1 and pod2 registration")
	}
}
