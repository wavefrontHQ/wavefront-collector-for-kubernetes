package discovery

import (
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	apicorev1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

func TestNotifyOfChanges(t *testing.T) {
	t.Run("calls notify", func(t *testing.T) {
		stopCh := make(chan struct{})
		discoveryCount := 0 * time.Second
		get := func() discovery.Config {
			discoveryCount++
			return discovery.Config{
				DiscoveryInterval: discoveryCount * time.Second,
			}
		}

		testSucceeded := false
		notify := func() {
			testSucceeded = true
			close(stopCh)
		}

		NotifyOfChanges(get, notify, 3*time.Second, stopCh)

		assert.True(t, testSucceeded)
	})

	t.Run("only calls notify when it gets a new value from the get function", func(t *testing.T) {
		stopCh := make(chan struct{})
		get := func() discovery.Config {
			return discovery.Config{
				DiscoveryInterval: 0,
			}
		}

		testSucceeded := false
		notify := func() {
			testSucceeded = true
			close(stopCh)
		}

		time.AfterFunc(15*time.Second, func() {
			close(stopCh)
		})
		NotifyOfChanges(get, notify, 3*time.Second, stopCh)

		assert.False(t, testSucceeded)
	})

	t.Run("retries until successful", func(t *testing.T) {
		stopCh := make(chan struct{})
		discoveryCount := 0 * time.Second
		get := func() discovery.Config {
			discoveryCount++
			if discoveryCount >= 4 {
				return discovery.Config{
					DiscoveryInterval: 1,
				}
			}

			return discovery.Config{
				DiscoveryInterval: 0,
			}
		}

		testSucceeded := false
		notify := func() {
			testSucceeded = true
			close(stopCh)
		}

		NotifyOfChanges(get, notify, 3*time.Second, stopCh)

		assert.True(t, testSucceeded)
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
