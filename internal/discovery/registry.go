package discovery

import (
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"k8s.io/apimachinery/pkg/util/wait"
)

type defaultRegistry struct {
	count   metrics.Gauge
	mtx     sync.RWMutex
	targets map[string]TargetHandler
}

func NewRegistry(name string) TargetRegistry {
	registry := &defaultRegistry{
		targets: make(map[string]TargetHandler),
		count:   metrics.GetOrRegisterGauge("discovery."+name+".targets.registered", metrics.DefaultRegistry),
	}

	// update the target counter once a minute
	go wait.Forever(func() {
		registry.count.Update(int64(registry.Count()))
	}, 1*time.Minute)

	return registry
}

func (registry *defaultRegistry) Register(name string, th TargetHandler) {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()
	registry.targets[name] = th
}

func (registry *defaultRegistry) Unregister(name string) {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()
	delete(registry.targets, name)
}

func (registry *defaultRegistry) Handler(name string) TargetHandler {
	registry.mtx.RLock()
	defer registry.mtx.RUnlock()
	return registry.targets[name]
}

func (registry *defaultRegistry) Encoding(name string) string {
	handler := registry.Handler(name)
	if handler != nil {
		return handler.Encoding(name)
	}
	return ""
}

func (registry *defaultRegistry) Count() int {
	registry.mtx.RLock()
	defer registry.mtx.RUnlock()
	return len(registry.targets)
}
