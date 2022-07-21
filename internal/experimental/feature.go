package experimental

const ClusterSource = "cluster-source"

var enabled = map[string]bool{}

func IsEnabled(name string) bool {
	return enabled[name]
}

func EnableFeature(name string) {
	enabled[name] = true
}

func DisableFeature(name string) {
	delete(enabled, name)
}
