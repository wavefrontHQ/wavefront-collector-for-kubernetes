package util

import (
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

type Incrementer interface {
	Inc(int642 int64)
}

func FilterAppend(filter filter.Filter, filtered Incrementer, points []*wf.Point, point *wf.Point) []*wf.Point {
	if filter == nil {
		return append(points, point)
	}
	if !filter.MatchMetric(point.Metric, point.Tags()) {
		log.WithField("name", point.Metric).Tracef("dropping metric")
		filtered.Inc(1)
		return points
	}
	point.FilterTags(filter.MatchTag)
	return append(points, point)
}
