package internal

import "sync/atomic"

// counter for internal metrics
type MetricCounter struct {
	value int64
}

func (c *MetricCounter) inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *MetricCounter) count() int64 {
	return atomic.LoadInt64(&c.value)
}

// functional gauge for internal metrics
type FunctionalGauge struct {
	value func() int64
}

func (g *FunctionalGauge) instantValue() int64 {
	return g.value()
}
