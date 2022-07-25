package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

type Diff struct {
	Missing  []*Metric
	Extra    []*Metric
	Unwanted []*Metric
}

func DiffMetrics(expected, excluded, actual []*Metric) *Diff {
	expectedKeyers := metricKeyers(expected)
	expectedKeyMap := metricKeyMap(expected, expectedKeyers)
	actualKeyMap := metricKeyMap(actual, expectedKeyers)
	missing, extra := disjunct(expectedKeyMap, actualKeyMap)

	excludedKeyers := metricKeyers(excluded)
	excludedKeyMap := metricKeyMap(excluded, excludedKeyers)
	actualExcludedKeyMap := metricKeyMap(actual, excludedKeyers)
	unwanted := intersect(excludedKeyMap, actualExcludedKeyMap)

	return &Diff{
		Missing:  missing,
		Extra:    extra,
		Unwanted: unwanted,
	}
}

// keyer returns whether or not it could generate a key and the key of the given metric
type keyer func(*Metric) (bool, string)

func metricKeyers(expected []*Metric) map[string][]keyer {
	keyersByMetric := map[string][]keyer{}
	for _, m := range expected {
		keyersByMetric[m.Name] = append(keyersByMetric[m.Name], metricKeyer(m))
	}
	return keyersByMetric
}

func metricKeyer(m *Metric) keyer {
	var keyers []keyer
	keyers = append(keyers, nameKey(m.Name))
	if m.Value != "" {
		keyers = append(keyers, valueKey(m.Value))
	}
	if m.Timestamp != "" {
		keyers = append(keyers, timestampKey(m.Timestamp))
	}
	keyers = append(keyers, tagsKey(m.Tags))
	return compositeKey(keyers...)
}

func compositeKey(keyers ...keyer) keyer {
	return func(metric *Metric) (bool, string) {
		var keys []string
		for _, keyer := range keyers {
			matched, key := keyer(metric)
			if !matched {
				return false, ""
			}
			keys = append(keys, key)
		}
		return true, strings.Join(keys, " ")
	}
}

func nameKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Name == expected, metric.Name
	}
}

func valueKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Value == expected, metric.Value
	}
}

func timestampKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Timestamp == expected, metric.Timestamp
	}
}

func tagNameKey(name string) keyer {
	key := strings.TrimPrefix(name, "!")
	if strings.HasPrefix(name, "!") {
		return func(metric *Metric) (bool, string) {
			_, exists := metric.Tags[key]
			return !exists, fmt.Sprintf("%s!=*", key)
		}
	}
	return func(metric *Metric) (bool, string) {
		_, exists := metric.Tags[key]
		return exists, fmt.Sprintf("%s=*", key)
	}
}

func fullTagKey(name, value string) keyer {
	key := strings.TrimPrefix(name, "!")
	if strings.HasPrefix(name, "!") {
		return func(metric *Metric) (bool, string) {
			var adjustedKey string
			if len(metric.Tags[key]) > 0 {
				adjustedKey = key
			} else {
				adjustedKey = strings.Replace(key, "!", "", 1)
			}
			return metric.Tags[adjustedKey] != value, fmt.Sprintf("%s!=%#v", key, metric.Tags[key])
		}
	}
	if strings.HasPrefix(name, "~") {
		g := glob.MustCompile(value)
		return func(metric *Metric) (bool, string) {
			var adjustedKey string
			if len(metric.Tags[key]) > 0 {
				adjustedKey = key
			} else {
				adjustedKey = strings.Replace(key, "~", "", 1)
			}
			return g.Match(metric.Tags[adjustedKey]), fmt.Sprintf("%s≅%#v", key, value)
		}
	}
	return func(metric *Metric) (bool, string) {
		return metric.Tags[key] == value, fmt.Sprintf("%s=%#v", key, metric.Tags[key])
	}
}

func tagsKey(tags map[string]string) keyer {
	tagNames := make([]string, 0, len(tags))
	for name := range tags {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)
	keyers := make([]keyer, len(tags))
	for i, name := range tagNames {
		if tags[name] == "" {
			keyers[i] = tagNameKey(name)
		} else {
			keyers[i] = fullTagKey(name, tags[name])
		}
	}
	return compositeKey(keyers...)
}

func metricKeyMap(metrics []*Metric, keyers map[string][]keyer) map[string]*Metric {
	keyMap := map[string]*Metric{}
	for _, metric := range metrics {
		foundKeyers := keyers[metric.Name]
		found := false
		for _, foundKeyer := range foundKeyers {
			matched, key := foundKeyer(metric)
			if matched {
				keyMap[key] = metric
				found = true
				break
			}
		}
		if !found {
			_, key := metricKeyer(metric)(metric)
			keyMap[key] = metric
		}
	}
	return keyMap
}

func disjunct(a, b map[string]*Metric) (onlyInA []*Metric, onlyInB []*Metric) {
	onlyInA = []*Metric{}
	onlyInB = []*Metric{}
	for x := range a {
		if _, exists := b[x]; !exists {
			onlyInA = append(onlyInA, a[x])
		}
	}
	for y := range b {
		if _, exists := a[y]; !exists {
			onlyInB = append(onlyInB, b[y])
		}
	}
	return onlyInA, onlyInB
}

func intersect(a, b map[string]*Metric) (common []*Metric) {
	common = []*Metric{}
	for x := range a {
		if v, exists := b[x]; exists {
			common = append(common, v)
		}
	}
	return common
}
