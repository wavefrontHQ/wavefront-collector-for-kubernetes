package prometheus

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type targetRegistry struct {
	mtx     sync.RWMutex
	targets map[string]*targetHandler
}

var (
	// singleton target registry
	registry    *targetRegistry
	targetCount metrics.Gauge
)

func init() {
	registry = &targetRegistry{
		targets: make(map[string]*targetHandler),
	}
	targetCount = metrics.GetOrRegisterGauge("discovery.targets.registered", metrics.DefaultRegistry)

	// update the target counter once a minute
	go wait.Forever(func() {
		targetCount.Update(int64(registry.count()))
	}, 1*time.Minute)
}

func (d *targetRegistry) register(name string, th *targetHandler) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.targets[name] = th
}

func (d *targetRegistry) unregister(name string) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	delete(d.targets, name)
}

func (d *targetRegistry) registered(name string) *targetHandler {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.targets[name]
}

func (d *targetRegistry) registeredURL(name string) string {
	handler := d.registered(name)
	if handler != nil {
		return handler.get(name)
	}
	return ""
}

func (d *targetRegistry) count() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return len(d.targets)
}
