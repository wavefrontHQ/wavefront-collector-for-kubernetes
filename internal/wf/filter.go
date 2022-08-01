package wf

import (
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

type Incrementer interface {
	Inc(int642 int64)
}

// FilterAppend appends the point to points when Filter does not return nil
func FilterAppend(filter filter.Filter, filtered Incrementer, points []Metric, point Metric) []Metric {
	point = Filter(filter, filtered, point)
	if point == nil {
		return points
	}
	return append(points, point)
}

// Filter returns nil when it does not match the supplied filter.Filter.
// Filter increments the Incrementor when filtering.
// Filter filters the tags on a matched point.
func Filter(filter filter.Filter, filtered Incrementer, point Metric) Metric {
	if filter == nil || point == nil {
		return point
	}
	if !filter.MatchMetric(point.Name(), point.Tags()) {
		log.WithField("name", point.Name()).Tracef("dropping metric")
		filtered.Inc(1)
		return nil
	}
	point.FilterTags(filter.MatchTag)
	return point
}
