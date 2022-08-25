package wf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/caio/go-tdigest"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"golang.org/x/crypto/blake2b"
)

type Centroid struct {
	Value float64
	Count float64
}

type Distribution struct {
	Cumulative bool
	name       string
	tags       map[string]string // TODO string interning
	Source     string            // TODO string interning
	Centroids  []Centroid
	Digest     *tdigest.TDigest
	Timestamp  time.Time
}

// NewCumulativeDistribution encodes prometheus style distribution.
func NewCumulativeDistribution(name string, source string, tags map[string]string, centroids []Centroid, timestamp time.Time) *Distribution {
	return newDistribution(true, name, source, tags, centroids, nil, timestamp)
}

// NewFrequencyDistribution encodes a WF style distribution.
func NewFrequencyDistribution(name string, source string, tags map[string]string, centroids []Centroid, timestamp time.Time) *Distribution {
	return newDistribution(false, name, source, tags, centroids, nil, timestamp)
}

func NewDigestDistribution(name string, source string, tags map[string]string, digest *tdigest.TDigest, timestamp time.Time) *Distribution {
	return newDistribution(false, name, source, tags, nil, digest, timestamp)
}

func newDistribution(cumulative bool, name string, source string, tags map[string]string, centroids []Centroid, digest *tdigest.TDigest, timestamp time.Time) *Distribution {
	sort.Slice(centroids, func(i, j int) bool {
		return centroids[i].Value < centroids[j].Value
	})
	return &Distribution{
		Cumulative: cumulative,
		name:       name,
		Source:     source,
		tags:       tags,
		Centroids:  centroids,
		Digest:     digest,
		Timestamp:  timestamp,
	}
}

func (d *Distribution) Clone() *Distribution {
	return newDistribution(d.Cumulative, d.Name(), d.Source, d.clonedTags(), d.clonedCentroids(), d.Digest, d.Timestamp)
}

func (d *Distribution) clonedTags() map[string]string {
	clonedTags := make(map[string]string, len(d.Tags()))
	for k, v := range d.Tags() {
		clonedTags[k] = v
	}
	return clonedTags
}

func (d *Distribution) clonedCentroids() []Centroid {
	cloned := make([]Centroid, len(d.Centroids))
	copy(cloned, d.Centroids)
	return cloned
}

func (d *Distribution) Points() int {
	return 7
}

func (d *Distribution) ToFrequency() *Distribution {
	if !d.Cumulative {
		return d
	}
	return NewDigestDistribution(
		d.Name(),
		d.Source,
		d.Tags(),
		smoothCentroids(deriveCentroids(d.Centroids)),
		d.Timestamp,
	)
}

func smoothCentroids(derivedCentroids []Centroid) *tdigest.TDigest {
	if len(derivedCentroids) == 1 && derivedCentroids[0].Value == math.Inf(1) {
		return nil
	}
	amplification := math.Max(1, 1/minCount(derivedCentroids))
	digest, _ := tdigest.New(tdigest.Compression(3.2))
	//centroidCounts := map[float64]float64{}
	for i, centroid := range derivedCentroids {
		currentBucketBound := centroid.Value
		actualBucketCount := derivedCentroids[i].Count * amplification
		if currentBucketBound <= 0 || actualBucketCount == 0 {
			continue
		}
		actualBucketCount = math.Max(1.0, actualBucketCount)
		previousBucketBound := 0.0
		if currentBucketBound == math.Inf(1) {
			currentBucketBound = derivedCentroids[i-1].Value
		}
		if i > 0 {
			previousBucketBound = derivedCentroids[i-1].Value
		} else if currentBucketBound > 0 {
			previousBucketBound = 0
		} else {
			previousBucketBound = currentBucketBound
		}

		lowerCount := math.Trunc(actualBucketCount / 4)
		if lowerCount > 0 {
			_ = digest.AddWeighted(previousBucketBound, uint64(lowerCount))
		}
		middleCount := math.Trunc(actualBucketCount / 2)
		if middleCount > 0 {
			_ = digest.AddWeighted((currentBucketBound+previousBucketBound)/2, uint64(middleCount))
		}
		upperCount := math.Trunc(actualBucketCount - lowerCount - middleCount)
		if upperCount > 0 {
			_ = digest.AddWeighted(currentBucketBound, uint64(upperCount))
		}
	}
	//centroids := make([]Centroid, 0, len(centroidCounts))
	//for value, count := range centroidCounts {
	//	centroids = append(centroids, Centroid{Value: value, Count: count})
	//}
	return digest
}

func minCount(centroids []Centroid) float64 {
	minCount := math.MaxFloat64
	for _, centroid := range centroids {
		if centroid.Count > 0 && centroid.Count < minCount {
			minCount = centroid.Count
		}
	}
	return minCount
}

func deriveCentroids(centroids []Centroid) []Centroid {
	var derived []Centroid
	for i, centroid := range centroids {
		deltaCount := 0.0
		if i > 0 {
			deltaCount = centroid.Count - centroids[i-1].Count
		} else {
			deltaCount = centroid.Count
		}
		derived = append(derived, Centroid{
			Value: centroid.Value,
			Count: deltaCount,
		})
	}
	return derived
}

func (d *Distribution) OverrideTag(name, value string) {
	if d == nil {
		return
	}
	if d.tags == nil {
		d.tags = map[string]string{}
	}
	d.tags[name] = value
}

func (d *Distribution) AddTags(tags map[string]string) {
	if tags == nil {
		return
	}
	for name, value := range tags {
		d.addTag(name, value)
	}
}

func (d *Distribution) addTag(name, value string) {
	if d == nil {
		return
	}
	if d.tags == nil {
		d.tags = map[string]string{}
	}
	if _, exists := d.tags[name]; !exists {
		d.tags[name] = value
	}
}

func (d *Distribution) SetSource(source string) {
	d.Source = source
}

func (d *Distribution) Name() string {
	return d.name
}

func (d *Distribution) Tags() map[string]string {
	return d.tags
}

func (d *Distribution) FilterTags(pred func(string) bool) {
	for name := range d.tags {
		if !pred(name) {
			delete(d.tags, name)
		}
	}
}

func (d *Distribution) Send(to Sender) error {
	if d.Cumulative {
		return errors.New("cannot send prometheus style distribution to wavefront")
	}
	var centroids []histogram.Centroid
	//centroids := wfCentroids(d)
	//if len(centroids) == 0 {
	//	return nil
	//}
	d.Digest.ForEachCentroid(func(mean float64, count uint64) bool {
		if count > 0 {
			centroids = append(centroids, histogram.Centroid{Value: mean, Count: int(count)})
		}
		return true
	})
	return to.SendDistribution(
		d.Name(),
		centroids,
		map[histogram.Granularity]bool{histogram.MINUTE: true},
		d.Timestamp.Unix(),
		d.Source,
		d.Tags(),
	)
}

func wfCentroids(d *Distribution) []histogram.Centroid {
	wfCentroids := make([]histogram.Centroid, 0, len(d.Centroids))
	for _, centroid := range d.Centroids {
		if centroid.Count == 0.0 {
			continue
		}
		wfCentroids = append(wfCentroids, histogram.Centroid{
			Value: centroid.Value,
			Count: int(centroid.Count),
		})
	}
	return wfCentroids
}

func (d *Distribution) Equals(b *Distribution) bool {
	return false
}

type DistributionHash [32]byte

func (d *Distribution) Key() DistributionHash {
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte(fmt.Sprintf("%t ", d.Cumulative)))
	buf.WriteString(d.name)
	buf.WriteString(" source=")
	buf.WriteString(d.Source)
	tags := d.Tags()
	var names []string
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		buf.WriteByte(' ')
		buf.WriteString(name)
		buf.WriteByte('=')
		buf.WriteString(tags[name])
	}
	return blake2b.Sum256(buf.Bytes())
}

func (d *Distribution) Rate(prev *Distribution) *Distribution {
	if prev == nil || prev.Key() != d.Key() {
		return nil
	}
	centroidRate := CentroidRate(d.Centroids, prev.Centroids, d.Timestamp.Sub(prev.Timestamp))
	if centroidRate == nil {
		return nil
	}
	return newDistribution(d.Cumulative, d.Name(), d.Source, d.Tags(), centroidRate, d.Digest, d.Timestamp)
}

func CentroidRate(curr, prev []Centroid, duration time.Duration) []Centroid {
	if len(curr) != len(prev) {
		return nil
	}
	var centroids = make([]Centroid, len(curr))
	for i := 0; i < len(curr); i++ {
		if curr[i].Value != prev[i].Value || curr[i].Count < prev[i].Count || duration.Seconds() <= 0 {
			return nil
		}
		centroids[i] = Centroid{
			Value: curr[i].Value,
			Count: (curr[i].Count - prev[i].Count) / duration.Minutes(),
		}
	}
	return centroids
}
