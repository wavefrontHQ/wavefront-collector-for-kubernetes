package metrics

import "sync"

type MetricStore struct {
	metrics     []*Metric
	metricsMu   *sync.Mutex
	badMetrics  []string
	badMetricMu *sync.Mutex
}

func NewMetricStore() *MetricStore {
	return &MetricStore{
		metrics:     make([]*Metric, 0, 1024),
		metricsMu:   &sync.Mutex{},
		badMetrics:  make([]string, 0, 1024),
		badMetricMu: &sync.Mutex{},
	}
}

func (s *MetricStore) Metrics() []*Metric {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	cpy := make([]*Metric, len(s.metrics))
	copy(cpy, s.metrics)
	return cpy
}

func (s *MetricStore) BadMetrics() []string {
	s.badMetricMu.Lock()
	defer s.badMetricMu.Unlock()
	cpy := make([]string, len(s.badMetrics))
	copy(cpy, s.badMetrics)
	return cpy
}

func (s *MetricStore) LogMetric(metric *Metric) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	s.metrics = append(s.metrics, metric)
}

func (s *MetricStore) LogBadMetric(metric string) {
	s.badMetricMu.Lock()
	defer s.badMetricMu.Unlock()
	s.badMetrics = append(s.badMetrics, metric)
}
