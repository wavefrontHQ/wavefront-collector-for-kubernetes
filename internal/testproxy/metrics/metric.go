package metrics

type Metric struct {
	Name      string
	Value     string
	Timestamp string
	Tags      map[string]string
}

func ParseMetric(line string) (*Metric, error) {
	g := &MetricGrammar{Buffer: line}
	g.Init()
	if err := g.Parse(); err != nil {
		return nil, err
	}
	g.Execute()
	if g.Histogram {
		return nil, nil
	}
	return &Metric{
		Name:      g.Name,
		Value:     g.Value,
		Timestamp: g.Timestamp,
		Tags:      g.Tags,
	}, nil
}
