package wf

type LabelPair struct {
	Name  *string
	Value *string
}

// Point is a single point in Wavefront metric format.
type Point struct {
	Metric    string
	Value     float64
	Timestamp int64
	Source    string

	tags       map[string]string
	labelPairs []LabelPair
}

func NewPoint(metric string, value float64, timestamp int64, source string, tags map[string]string) *Point {
	return &Point{
		Metric:    metric,
		Value:     value,
		Timestamp: timestamp,
		Source:    source,
		tags:      tags,
	}
}

func (m *Point) SetLabelPairs(pairs []LabelPair) {
	m.labelPairs = pairs
}

func (m *Point) SetTags(tags map[string]string) {
	m.tags = tags
}

// OverrideTag sets a tag regardless of whether it already exists
func (m *Point) OverrideTag(name, value string) {
	if m == nil {
		return
	}
	if m.tags == nil {
		m.tags = map[string]string{}
	}
	m.tags[name] = value
}

// AddTag adds a tag if it does not already exist
func (m *Point) AddTag(name, value string) {
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
func (m *Point) AddTags(tags map[string]string) {
	if tags == nil {
		return
	}
	for name, value := range tags {
		m.AddTag(name, value)
	}
}

func (m *Point) Tags() map[string]string {
	tags := make(map[string]string, len(m.labelPairs))
	for _, labelPair := range m.labelPairs {
		tags[*labelPair.Name] = *labelPair.Value
	}

	for k, v := range m.tags {
		tags[k] = v
	}

	return tags
}

func (m *Point) FilterTags(pred func(string) bool) {
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
