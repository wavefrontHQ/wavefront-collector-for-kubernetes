package experimental

import "sync"

const ClusterSource = "cluster-source"

var (
	mu      sync.RWMutex
	enabled = map[string]bool{}
)

func IsEnabled(name string) bool {
	mu.RLock()
	isEnabled := enabled[name]
	mu.RUnlock()
	return isEnabled
}

func EnableFeature(name string) {
	mu.Lock()
	enabled[name] = true
	mu.Unlock()
}

func DisableFeature(name string) {
	mu.Lock()
	delete(enabled, name)
	mu.Unlock()
}

func DisableAll() {
	mu.Lock()
	enabled = map[string]bool{}
	mu.Unlock()
}
