package wf

// Distribution is single HS in Wavefront distribution format.
type Distribution struct {
	Metric    string
	Centroids []Centroid
	Timestamp int64
	Source    string

	tags       map[string]string
	labelPairs []LabelPair
}

// Centroid encapsulates a mean value and the count of points associated with that value.
type Centroid struct {
	Value float64
	Count int
}

func NewDistribution(metric string, centroids []Centroid, timestamp int64, source string, tags map[string]string) *Distribution {
	return &Distribution{
		Metric:    metric,
		Centroids: centroids,
		Timestamp: timestamp,
		Source:    source,
		tags:      tags,
	}
}

func (m *Distribution) SetLabelPairs(pairs []LabelPair) {
	m.labelPairs = pairs
}

func (m *Distribution) SetTags(tags map[string]string) {
	m.tags = tags
}

// OverrideTag sets a tag regardless of whether it already exists
func (m *Distribution) OverrideTag(name, value string) {
	if m == nil {
		return
	}
	if m.tags == nil {
		m.tags = map[string]string{}
	}
	m.tags[name] = value
}

// AddTag adds a tag if it does not already exist
func (m *Distribution) AddTag(name, value string) {
	if m == nil {
		return
	}
	if m.tags == nil {
		m.tags = map[string]string{}
	}
	if _, exists := m.tags[name]; !exists {
		m.tags[name] = value
	}
}

// AddTags adds any tags that do not already exist
func (m *Distribution) AddTags(tags map[string]string) {
	if tags == nil {
		return
	}
	for name, value := range tags {
		m.AddTag(name, value)
	}
}

func (m *Distribution) Tags() map[string]string {
	tags := make(map[string]string, len(m.labelPairs))
	for _, labelPair := range m.labelPairs {
		tags[*labelPair.Name] = *labelPair.Value
	}

	for k, v := range m.tags {
		tags[k] = v
	}

	return tags
}

func (m *Distribution) FilterTags(pred func(string) bool) {
	var nextLabelPairs []LabelPair
	for _, labelPair := range m.labelPairs {
		if pred(*labelPair.Name) {
			nextLabelPairs = append(nextLabelPairs, labelPair)
		}
	}
	m.labelPairs = nextLabelPairs

	for name := range m.tags {
		if !pred(name) {
			delete(m.tags, name)
		}
	}
}
