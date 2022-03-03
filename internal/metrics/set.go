package metrics

import "time"

// Set is a collection of metrics tied to a specific resource
type Set struct {
	// CollectionStartTime is a time since when the metrics are collected for this entity.
	// It is affected by events like entity (e.g. pod) creation, entity restart (e.g. for container),
	// Kubelet restart.
	CollectionStartTime time.Time
	// EntityCreateTime is a time of entity creation and persists through entity restarts and
	// Kubelet restarts.
	EntityCreateTime time.Time
	ScrapeTime       time.Time
	Values           map[string]Value
	Labels           map[string]string
	LabeledValues    []LabeledValue
}

// FindLabels returns the labels for a given metric name
func (s *Set) FindLabels(name string) (map[string]string, bool) {
	_, found := s.Values[name]
	if found {
		return s.Labels, true
	}
	for _, labeledValue := range s.LabeledValues {
		if labeledValue.Name == name {
            labels := make(map[string]string, len(s.Labels) + len(labeledValue.Labels))
            for k, v := range labeledValue.Labels {
                labels[k] = v
            }
            for k, v := range s.Labels {
                labels[k] = v
            }
			return labels, true
		}
	}
	return map[string]string{}, false
}
