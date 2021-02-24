package reporting

import (
	"fmt"
	"reflect"

	metrics "github.com/rcrowley/go-metrics"
)

// RegisterMetric tag support for metrics.Register()
// return RegistryError if the metric is not registered
//
// Deprecated: Use WavefrontMetricsReporter.RegisterMetric instead.
func RegisterMetric(name string, metric interface{}, tags map[string]string) error {
	key := EncodeKey(name, tags)
	err := metrics.Register(key, metric)
	if err != nil {
		return err
	}
	m := GetMetric(name, tags)
	if m == nil {
		return RegistryError(fmt.Sprintf("Metric '%s'(%s) not registered.", name, reflect.TypeOf(metric).String()))
	}
	return nil
}

// GetMetric tag support for metrics.Get()
//
// Deprecated: Use WavefrontMetricsReporter.GetMetric instead.
func GetMetric(name string, tags map[string]string) interface{} {
	key := EncodeKey(name, tags)
	return metrics.Get(key)
}

// GetOrRegisterMetric tag support for metrics.GetOrRegister()
//
// Deprecated: Use WavefrontMetricsReporter.GetOrRegisterMetric instead.
func GetOrRegisterMetric(name string, i interface{}, tags map[string]string) interface{} {
	key := EncodeKey(name, tags)
	return metrics.GetOrRegister(key, i)
}

// UnregisterMetric tag support for metrics.UnregisterMetric()
//
// Deprecated: Use WavefrontMetricsReporter.UnregisterMetric instead.
func UnregisterMetric(name string, tags map[string]string) {
	key := EncodeKey(name, tags)
	metrics.Unregister(key)
}
