package filter

import (
	"fmt"
	"strings"
)

func FromQuery(vals map[string][]string) Config {
	if len(vals) == 0 {
		return Config{}
	}

	metricWhitelist := vals[MetricWhitelist]
	metricBlacklist := vals[MetricBlacklist]
	metricTagWhitelist := parseFilters(vals[MetricTagWhitelist])
	metricTagBlacklist := parseFilters(vals[MetricTagBlacklist])
	tagInclude := vals[TagInclude]
	tagExclude := vals[TagExclude]

	if len(metricWhitelist) == 0 && len(metricBlacklist) == 0 && len(metricTagWhitelist) == 0 &&
		len(metricTagBlacklist) == 0 && len(tagInclude) == 0 && len(tagExclude) == 0 {
		return Config{}
	}

	return Config{
		MetricWhitelist:    metricWhitelist,
		MetricBlacklist:    metricBlacklist,
		MetricTagWhitelist: metricTagWhitelist,
		MetricTagBlacklist: metricTagBlacklist,
		TagInclude:         tagInclude,
		TagExclude:         tagExclude,
	}
}

func FromConfig(cfg Config) Filter {
	if cfg.Empty() {
		return nil
	}

	metricWhitelist := cfg.MetricWhitelist
	metricBlacklist := cfg.MetricBlacklist
	metricTagWhitelist := cfg.MetricTagWhitelist
	metricTagBlacklist := cfg.MetricTagBlacklist
	tagInclude := cfg.TagInclude
	tagExclude := cfg.TagExclude

	if len(metricWhitelist) == 0 && len(metricBlacklist) == 0 && len(metricTagWhitelist) == 0 &&
		len(metricTagBlacklist) == 0 && len(tagInclude) == 0 && len(tagExclude) == 0 {
		return nil
	}

	return NewGlobFilter(Config{
		MetricWhitelist:    metricWhitelist,
		MetricBlacklist:    metricBlacklist,
		MetricTagWhitelist: metricTagWhitelist,
		MetricTagBlacklist: metricTagBlacklist,
		TagInclude:         tagInclude,
		TagExclude:         tagExclude,
	})
}

func parseFilters(slice []string) map[string][]string {
	if len(slice) == 0 {
		return nil
	}
	out := make(map[string][]string)

	// each string in the slice is of the form: "tagK:[glob1, glob2, ...]"
	for _, tag := range slice {
		s := strings.Split(tag, ":")
		if len(s) == 2 {
			k, v := s[0], s[1]
			patterns, err := parseValue(v)
			if err != nil {
				fmt.Print(err)
			} else {
				out[k] = patterns
			}
		}
	}
	return out
}

// Gets a string slice from a string of the form "[foo*, bar*, ...]"
func parseValue(val string) ([]string, error) {
	if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
		return nil, fmt.Errorf("invalid metric tag filter: %s", val)
	}
	tagValue := val[1 : len(val)-1]
	return strings.Split(tagValue, ","), nil
}
