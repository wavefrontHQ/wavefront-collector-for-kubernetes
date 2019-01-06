package prometheus

import "sync"

type targetRegistry struct {
	mtx     sync.RWMutex
	targets map[string]*targetHandler
}

// singleton target registry
var registry *targetRegistry

func init() {
	registry = &targetRegistry{
		targets: make(map[string]*targetHandler),
	}
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
		return handler.get(name).scrapeURL
	}
	return ""
}
