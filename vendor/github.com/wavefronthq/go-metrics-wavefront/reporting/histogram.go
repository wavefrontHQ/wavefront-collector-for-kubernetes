package reporting

import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

// Histogram wrapper of Wavefront Histogram so it can be used with metrics.Registry
type Histogram struct {
	delegate histogram.Histogram
}

// NewHistogram create a new Wavefront Histogram and the wrapper
func NewHistogram(options ...histogram.Option) metrics.Histogram {
	return Histogram{delegate: histogram.New(options...)}
}

// Clear will panic
func (h Histogram) Clear() {
	panic("Clear called on a Histogram")
}

// Count returns the total number of samples on this histogram.
func (h Histogram) Count() int64 {
	return int64(h.delegate.Count())
}

// Min returns the minimum Value of samples on this histogram.
func (h Histogram) Min() int64 {
	return int64(h.delegate.Min())
}

// Max returns the maximum Value of samples on this histogram.
func (h Histogram) Max() int64 {
	return int64(h.delegate.Max())
}

// Sum returns the sum of all values on this histogram.
func (h Histogram) Sum() int64 {
	return int64(h.delegate.Sum())
}

// Mean returns the mean values of samples on this histogram.
func (h Histogram) Mean() float64 {
	return h.delegate.Mean()
}

// Update registers a new sample in the histogram.
func (h Histogram) Update(v int64) {
	h.delegate.Update(float64(v))
}

// Sample will panic
func (h Histogram) Sample() metrics.Sample {
	panic("Sample not supported")
}

// Snapshot create a metrics.Histogram
func (h Histogram) Snapshot() metrics.Histogram {
	c := 0
	for _, distribution := range h.delegate.Snapshot() {
		for _, centroid := range distribution.Centroids {
			c += centroid.Count
		}
	}

	sample := metrics.NewUniformSample(c)
	for _, distribution := range h.delegate.Snapshot() {
		for _, centroid := range distribution.Centroids {
			for i := 0; i < centroid.Count; i++ {
				sample.Update(int64(centroid.Value))
			}
		}
	}
	return metrics.NewHistogram(sample)
}

// StdDev returns the standard deviation.
func (h Histogram) StdDev() float64 {
	return h.Snapshot().StdDev()
}

// Variance returns the variance of inputs.
func (h Histogram) Variance() float64 {
	return h.Snapshot().Variance()
}

// Percentile returns the desired percentile estimation.
func (h Histogram) Percentile(p float64) float64 {
	return h.delegate.Quantile(p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the sample
func (h Histogram) Percentiles(ps []float64) []float64 {
	var res []float64
	for _, p := range ps {
		res = append(res, h.Percentile(p))
	}
	return res
}

// Distributions returns all samples on completed time slices, and clear the histogram
func (h Histogram) Distributions() []histogram.Distribution {
	return h.delegate.Distributions()
}

// Granularity value
func (h Histogram) Granularity() histogram.Granularity {
	return h.delegate.Granularity()
}
