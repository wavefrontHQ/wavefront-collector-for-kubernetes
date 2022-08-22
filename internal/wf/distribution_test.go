package wf_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

func TestDistribution(t *testing.T) {
	t.Run("Key", func(t *testing.T) {
		original := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{}, time.Now())

		diffTagOrder := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"ctag": "cvalue", "atag": "avalue", "btag": "bvalue"}, []wf.Centroid{}, time.Now())
		assert.Equal(t, original.Key(), diffTagOrder.Key())

		diffCumulative := wf.NewFrequencyDistribution("name1", "source1", map[string]string{"ctag": "cvalue", "atag": "avalue", "btag": "bvalue"}, []wf.Centroid{}, time.Now())
		assert.NotEqual(t, original.Key(), diffCumulative.Key())

		nameDiff := wf.NewCumulativeDistribution("name2", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{}, time.Now())
		assert.NotEqual(t, original.Key(), nameDiff.Key())

		sourceDiff := wf.NewCumulativeDistribution("name1", "source2", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{}, time.Now())
		assert.NotEqual(t, original.Key(), sourceDiff.Key())

		tagValueDiff := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue2", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{}, time.Now())
		assert.NotEqual(t, original.Key(), tagValueDiff.Key())

		tagNameDiff := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag2": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{}, time.Now())
		assert.NotEqual(t, original.Key(), tagNameDiff.Key())
	})

	t.Run("Rate", func(t *testing.T) {
		t.Run("Calculate Centroid rate", func(t *testing.T) {
			prevTimeStamp := time.Now()
			prev := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 1, Count: 2}}, prevTimeStamp)
			current := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 1, Count: 3}}, prevTimeStamp.Add(time.Minute))

			currRate := current.Rate(prev)
			assert.NotNil(t, currRate)
			assert.Equal(t, current.Cumulative, currRate.Cumulative)
			assert.Equal(t, current.Name(), currRate.Name())
			assert.Equal(t, current.Source, currRate.Source)
			assert.Equal(t, current.Tags(), currRate.Tags())
			assert.Equal(t, current.Timestamp, currRate.Timestamp)
			assert.Equal(t, []wf.Centroid{{Value: 1, Count: 1}}, currRate.Centroids)
		})

		t.Run("Doesn't calculate rate on different distributions", func(t *testing.T) {
			prevTimeStamp := time.Now()
			prev := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 1, Count: 2}}, prevTimeStamp)
			current := wf.NewCumulativeDistribution("name2", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 1, Count: 70}}, prevTimeStamp.Add(time.Minute))
			currRate := current.Rate(prev)
			assert.Nil(t, currRate)
		})

		t.Run("Doesn't calculate rate for non compatible centroids", func(t *testing.T) {
			prevTimeStamp := time.Now()
			prev := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 1, Count: 2}}, prevTimeStamp)
			current := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 2, Count: 70}}, prevTimeStamp.Add(time.Minute))
			currRate := current.Rate(prev)
			assert.Nil(t, currRate)
		})

		t.Run("Doesn't send rate if previous is nil", func(t *testing.T) {
			prevTimeStamp := time.Now()
			current := wf.NewCumulativeDistribution("name1", "source1", map[string]string{"btag": "bvalue", "atag": "avalue", "ctag": "cvalue"}, []wf.Centroid{{Value: 2, Count: 70}}, prevTimeStamp.Add(time.Minute))
			currRate := current.Rate(nil)
			assert.Nil(t, currRate)
		})
	})

	t.Run("Coverts Cumulative to density distribution", func(t *testing.T) {
		density := wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{
				{Value: 0.05, Count: 24054},
				{Value: 0.1, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			time.Now(),
		).ToFrequency()

		assert.Equal(t, 11, len(density.Centroids))
		assertCentroids(t, []wf.Centroid{
			{Value: 0.0, Count: 6013},
			{Value: 0.025, Count: 12027},
			{Value: 0.05, Count: 8361},
			{Value: 0.075, Count: 4695},
			{Value: 0.1, Count: 19085},
			{Value: 0.15, Count: 33474},
			{Value: 0.2, Count: 23986},
			{Value: 0.35, Count: 14498},
			{Value: 0.5, Count: 8399},
			{Value: 0.75, Count: 2299},
			{Value: 1.0, Count: 11483},
		}, density.Centroids)
	})

	t.Run("Converts cumulative distribution with negative lower bounds", func(t *testing.T) {
		density := wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{
				{Value: -0.1, Count: 24054},
				{Value: -0.05, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			time.Now(),
		).ToFrequency()

		assert.Equal(t, 9, len(density.Centroids))
		assertCentroids(t, []wf.Centroid{
			{Value: -0.1, Count: float64(24054 + 9390/4)},
			{Value: -0.075, Count: float64(9390 / 2)},
			{Value: -0.05, Count: float64(9390 - 9390/4 - 9390/2 + 66948/4)},
			{Value: 0.075, Count: float64(66948 / 2)},
			{Value: 0.2, Count: float64(66948 - 66948/4 - 66948/2 + 28997/4)},
			{Value: 0.35, Count: float64(28997 / 2)},
			{Value: 0.5, Count: float64(28997 - 28997/4 - 28997/2 + 4599/4)},
			{Value: 0.75, Count: float64(4599 / 2)},
			{Value: 1.0, Count: float64(4599 - 4599/4 - 4599/2 + (144320 - 133988))},
		}, density.Centroids)
	})

	t.Run("Converts cumulative distribution with amplification", func(t *testing.T) {
		density := wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{
				{Value: -0.1, Count: 0.5},
				{Value: -0.05, Count: 1},
			},
			time.Now(),
		).ToFrequency()

		assert.Equal(t, 3, len(density.Centroids))
		assertCentroids(t, []wf.Centroid{
			{Value: -0.1, Count: 1},
			{Value: -0.075, Count: 0},
			{Value: -0.05, Count: 1},
		}, density.Centroids)
	})

	t.Run("Leaves non-cumulative distribution unchanged", func(t *testing.T) {
		d := wf.NewFrequencyDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{
				{Value: -0.1, Count: 0.5},
				{Value: -0.05, Count: 1},
			},
			time.Now(),
		)

		assert.Equal(t, d, d.ToFrequency())
	})

	t.Run("Sends to WF", func(t *testing.T) {
		ts := time.Now()
		expectedTags := map[string]string{"sometag": "somevalue"}
		d := wf.NewFrequencyDistribution(
			"some.distribution",
			"somesource",
			expectedTags,
			[]wf.Centroid{
				{Value: 0.05, Count: 24054},
				{Value: 0.1, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			ts,
		)
		sender := NewMockDistributionSender(
			"some.distribution",
			[]histogram.Centroid{
				{Value: 0.05, Count: 24054},
				{Value: 0.1, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			map[histogram.Granularity]bool{histogram.MINUTE: true},
			ts.Unix(),
			"somesource",
			expectedTags,
		)

		_ = d.Send(sender)

		sender.Verify(t)
	})

	t.Run("Does not Send centroids with zero count to WF", func(t *testing.T) {
		ts := time.Now()
		expectedTags := map[string]string{"sometag": "somevalue"}
		d := wf.NewFrequencyDistribution(
			"some.distribution",
			"somesource",
			expectedTags,
			[]wf.Centroid{
				{Value: 0.05, Count: 1},
				{Value: 0.1, Count: 0},
			},
			ts,
		)
		sender := NewMockDistributionSender(
			"some.distribution",
			[]histogram.Centroid{
				{Value: 0.05, Count: 1},
			},
			map[histogram.Granularity]bool{histogram.MINUTE: true},
			ts.Unix(),
			"somesource",
			expectedTags,
		)

		_ = d.Send(sender)

		sender.Verify(t)
	})

	t.Run("Does not Send empty distributions to WF", func(t *testing.T) {
		ts := time.Now()
		expectedTags := map[string]string{"sometag": "somevalue"}
		d := wf.NewFrequencyDistribution(
			"some.distribution",
			"somesource",
			expectedTags,
			[]wf.Centroid{},
			ts,
		)

		assert.NoError(t, d.Send(nil))
	})

	t.Run("Does not send cumulative distributions to WF", func(t *testing.T) {
		ts := time.Now()
		expectedTags := map[string]string{"sometag": "somevalue"}
		d := wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			expectedTags,
			[]wf.Centroid{
				{Value: 0.05, Count: 24054},
				{Value: 0.1, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			ts,
		)
		assert.Error(t, d.Send(nil), "cannot send prometheus style distribution to wavefront")
	})
}

type MockDistributionSender struct {
	expectedName      string
	expectedCentroids []histogram.Centroid
	expectedHGS       map[histogram.Granularity]bool
	expectedTS        int64
	expectedSource    string
	expectedTags      map[string]string

	actualName      string
	actualCentroids []histogram.Centroid
	actualHGS       map[histogram.Granularity]bool
	actualTS        int64
	actualSource    string
	actualTags      map[string]string
}

func NewMockDistributionSender(
	expectedName string,
	expectedCentroids []histogram.Centroid,
	expectedHGS map[histogram.Granularity]bool,
	expectedTS int64,
	expectedSource string,
	expectedTags map[string]string,
) *MockDistributionSender {
	return &MockDistributionSender{
		expectedName:      expectedName,
		expectedCentroids: expectedCentroids,
		expectedHGS:       expectedHGS,
		expectedTS:        expectedTS,
		expectedSource:    expectedSource,
		expectedTags:      expectedTags,
	}
}

func (m *MockDistributionSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	panic("should not call SendMetric")
}

func (m *MockDistributionSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	m.actualName = name
	m.actualCentroids = centroids
	m.actualHGS = hgs
	m.actualTS = ts
	m.actualSource = source
	m.actualTags = tags
	return nil
}

func (m *MockDistributionSender) Verify(t *testing.T) {
	assert.Equal(t, m.expectedName, m.actualName)
	assert.Equal(t, m.expectedCentroids, m.actualCentroids)
	assert.Equal(t, m.expectedHGS, m.actualHGS)
	assert.Equal(t, m.expectedTS, m.actualTS)
	assert.Equal(t, m.expectedSource, m.actualSource)
	assert.Equal(t, m.expectedTags, m.actualTags)
}

func assertCentroids(t *testing.T, expectedCentroids []wf.Centroid, actualCentroids []wf.Centroid) {
	for i, expectedCentroid := range expectedCentroids {
		assert.InDeltaf(t, expectedCentroid.Value, actualCentroids[i].Value, 0.000000000001, "values are close")
		assert.Equal(t, expectedCentroid.Count, actualCentroids[i].Count)
	}
}

func TestCentroidRate(t *testing.T) {
	t.Run("computes the rate when the centroids when the rate above 0", func(t *testing.T) {
		prev := []wf.Centroid{{Value: 1, Count: 2}}
		curr := []wf.Centroid{{Value: 1, Count: 5}}
		expected := []wf.Centroid{{Value: 1, Count: 1.5}}
		actual := wf.CentroidRate(curr, prev, 2.0*time.Minute)
		assert.Equal(t, expected, actual)
	})

	t.Run("computes for multiple centroids", func(t *testing.T) {
		prev := []wf.Centroid{{Value: 1, Count: 2}, {Value: 4, Count: 5}}
		curr := []wf.Centroid{{Value: 1, Count: 3}, {Value: 4, Count: 7}}
		expected := []wf.Centroid{{Value: 1, Count: 0.5}, {Value: 4, Count: 1}}
		actual := wf.CentroidRate(curr, prev, 2.0*time.Minute)
		assert.Equal(t, expected, actual)
	})

	t.Run("does not diff centroids that don't have the same length", func(t *testing.T) {
		prev := []wf.Centroid{{Value: 1, Count: 2}}
		curr := []wf.Centroid{{Value: 1, Count: 3}, {Value: 4, Count: 7}}
		actual := wf.CentroidRate(curr, prev, 1.0*time.Minute)
		assert.Nil(t, actual)
	})

	t.Run("does not diff centroids that have all values in common", func(t *testing.T) {
		prev := []wf.Centroid{{Value: 1, Count: 2}, {Value: 6, Count: 8}}
		curr := []wf.Centroid{{Value: 1, Count: 3}, {Value: 4, Count: 7}}
		actual := wf.CentroidRate(curr, prev, 1.0*time.Minute)
		assert.Nil(t, actual)
	})

	t.Run("return nil to create gap when counter resets (case when centroids have decrementing count)", func(t *testing.T) {
		prev := []wf.Centroid{{Value: 1, Count: 10}}
		curr := []wf.Centroid{{Value: 1, Count: 2}}
		actual := wf.CentroidRate(curr, prev, 1.0*time.Minute)
		assert.Nil(t, actual)
	})
}