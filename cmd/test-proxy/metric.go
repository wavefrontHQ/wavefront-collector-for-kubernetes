package main

import "regexp"

type Metric struct {
	Name      string
	Value     string
	Timestamp string
	Tags      map[string]string
}

func ParseMetric(line string) (*Metric, error) {
	// To remove the special character ∆ from the metrics ∆~sdk.go.core.*
	reg, err := regexp.Compile("[∆]+")
	if err != nil {
		return nil, err
	}
	line = reg.ReplaceAllString(line, "")

	g := &MetricGrammar{Buffer: line}
	g.Init()
	if err := g.Parse(); err != nil {
		return nil, err
	}
	g.Execute()
	return &Metric{
		Name:      g.Name,
		Value:     g.Value,
		Timestamp: g.Timestamp,
		Tags:      g.Tags,
	}, nil
}
