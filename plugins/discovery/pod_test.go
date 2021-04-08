package discovery

import (
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"testing"
)

type DummyDiscoverer struct{}

func (discoverer *DummyDiscoverer) Delete(resource discovery.Resource) {}

func (discoverer *DummyDiscoverer) DeleteAll() {}

func (discoverer *DummyDiscoverer) Stop() {}

func (discoverer *DummyDiscoverer) Discover(resource discovery.Resource) {}

func NewDummyDiscoverer() *DummyDiscoverer {
	return &DummyDiscoverer{}
}

func TestUpdatePodDoesntPanic(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "influxdb-v2",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	}

	fakeDiscoverer := NewDummyDiscoverer()

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: pod}

	assert.NotPanics(t, func() { updatePodIfValid(obj, fakeDiscoverer) }, "updatePodIfValid panicked")
}

func TestDeletePodDoesntPanic(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "influxdb-v2",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	}

	fakeDiscoverer := NewDummyDiscoverer()

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: pod}

	assert.NotPanics(t, func() { deletePodIfValid(obj, fakeDiscoverer) }, "deletePodIfValid panicked")
}
