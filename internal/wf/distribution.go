package wf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

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
	Timestamp  time.Time
}

// NewCumulativeDistribution encodes prometheus style distribution.
func NewCumulativeDistribution(name string, source string, tags map[string]string, centroids []Centroid, timestamp time.Time) *Distribution {
	return newDistribution(true, name, source, tags, centroids, timestamp)
}

// NewFrequencyDistribution encodes a WF style distribution.
func NewFrequencyDistribution(name string, source string, tags map[string]string, centroids []Centroid, timestamp time.Time) *Distribution {
	return newDistribution(false, name, source, tags, centroids, timestamp)
}

func newDistribution(cumulative bool, name string, source string, tags map[string]string, centroids []Centroid, timestamp time.Time) *Distribution {
	sort.Slice(centroids, func(i, j int) bool {
		return centroids[i].Value < centroids[j].Value
	})
	return &Distribution{
		Cumulative: cumulative,
		name:       name,
		Source:     source,
		tags:       tags,
		Centroids:  centroids,
		Timestamp:  timestamp,
	}
}

func (d *Distribution) Clone() *Distribution {
	return newDistribution(d.Cumulative, d.Name(), d.Source, d.clonedTags(), d.clonedCentroids(), d.Timestamp)
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
	return NewFrequencyDistribution(
		d.Name(),
		d.Source,
		d.Tags(),
		smoothCentroids(deriveCentroids(d.Centroids)),
		d.Timestamp,
	)
}

func smoothCentroids(derivedCentroids []Centroid) []Centroid {
	if len(derivedCentroids) == 0 || (len(derivedCentroids) == 1 && derivedCentroids[0].Value == math.Inf(1)) {
		return nil
	}
	amplification := math.Max(1, 1/minCount(derivedCentroids))
	centroidCounts := map[float64]float64{}
	for i, centroid := range derivedCentroids {
		currBound := centroid.Value
		currCount := derivedCentroids[i].Count * amplification
		if currBound <= 0 || currCount == 0 {
			continue
		}
		prevBound := 0.0
		if i > 0 {
			prevBound = derivedCentroids[i-1].Value
		} else if currBound > 0 {
			prevBound = 0
		} else {
			prevBound = currBound
		}
		if currBound == math.Inf(1) {
			currBound = prevBound
		}
		lowerCount := math.Trunc(currCount / 4)
		centroidCounts[prevBound] += lowerCount
		middleCount := math.Trunc(currCount / 2)
		centroidCounts[(currBound+prevBound)/2] += middleCount
		upperCount := math.Trunc(currCount - lowerCount - middleCount)
		centroidCounts[currBound] += upperCount
	}
	centroids := make([]Centroid, 0, 2*len(derivedCentroids)-1)
	for value, count := range centroidCounts {
		centroids = append(centroids, Centroid{Value: value, Count: count})
	}
	return centroids
}

func minCount(derivedCentroids []Centroid) float64 {
	minCount := math.MaxFloat64
	for _, centroid := range derivedCentroids {
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
			Count: deltaCount,
			Value: centroid.Value,
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
	centroids := wfCentroids(d)
	if len(centroids) == 0 {
		return nil
	}
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
	return newDistribution(d.Cumulative, d.Name(), d.Source, d.Tags(), centroidRate, d.Timestamp)
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
