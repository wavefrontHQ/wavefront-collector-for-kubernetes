package discovery

import (
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
    "k8s.io/client-go/kubernetes/fake"
)

func TestManager(t *testing.T) {
	t.Run("config resync", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset()
		providerHandler := &fakeProviderHandler{}
		discoveryManager := NewDiscoveryManager(RunConfig{
			KubeClient: kubeClient,
			Handler:    providerHandler,
			Lister: &stubPodLister{
				Pods: []*apicorev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my-pod",
						Labels: map[string]string{"please-discover-me": "true"},
					},
				}},
			},
		})
		discoveryManager.Start()

		assert.Equal(t, 0, len(providerHandler.SourceProviders))

		// add new configmaps to be discovered
		// "wait" for the resync loop to pick it up (we should probably just figure out a way to force the resync)

		assert.Equal(t, 1, len(providerHandler.SourceProviders))
	})
}

type fakeProviderHandler struct {
	SourceProviders []metrics.SourceProvider
}

func (f *fakeProviderHandler) AddProvider(provider metrics.SourceProvider) {
	f.SourceProviders = append(f.SourceProviders, provider)
}

func (f *fakeProviderHandler) DeleteProvider(name string) {
	filteredProviders := make([]metrics.SourceProvider, 0, len(f.SourceProviders)-1)
	for _, provider := range f.SourceProviders {
		if provider.Name() == name {
			continue
		}
		filteredProviders = append(filteredProviders, provider)
	}
	f.SourceProviders = filteredProviders
}

type stubPodLister struct {
	Pods []*apicorev1.Pod
	Err  error
}

func (s *stubPodLister) ListPods(ns string, labels map[string]string) ([]*apicorev1.Pod, error) {
	return s.Pods, s.Err
}

func (s *stubPodLister) ListServices(ns string, labels map[string]string) ([]*apicorev1.Service, error) {
	return nil, nil
}

func (s *stubPodLister) ListNodes() ([]*apicorev1.Node, error) {
	return nil, nil
}
