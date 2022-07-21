package experimental

var features = map[string]bool{}

func IsEnabled(name string) bool {
	return features[name]
}

func EnableFeature(name string) {
	features[name] = true
}

func DisableFeature(name string) {
	delete(features, name)
}
