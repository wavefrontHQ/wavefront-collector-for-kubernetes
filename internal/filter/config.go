package filter

const (
	MetricWhitelist    = "metricWhitelist"
	MetricBlacklist    = "metricBlacklist"
	MetricTagWhitelist = "metricTagWhitelist"
	MetricTagBlacklist = "metricTagBlacklist"
	TagInclude         = "tagInclude"
	TagExclude         = "tagExclude"
)

// Configuration for filtering metrics.
// All the filtering options are applied at the end after specified prefixes etc are applied.
type Config struct {
	// List of glob pattern strings. Only metrics with names matching the whitelist are reported.
	MetricWhitelist []string `yaml:"metricWhitelist"`

	// List of glob pattern strings. Metrics with names matching the blacklist are dropped.
	MetricBlacklist []string `yaml:"metricBlacklist"`

	// List of glob pattern strings. Only metrics containing tag keys matching the whitelist will be reported.
	MetricTagWhitelist map[string][]string `yaml:"metricTagWhitelist"`

	// List of glob pattern strings. Metrics containing blacklisted tag keys will be dropped.
	MetricTagBlacklist map[string][]string `yaml:"metricTagBlacklist"`

	// List of glob pattern strings. Tags with matching keys will be included. All other tags will be excluded.
	TagInclude []string `yaml:"tagInclude"`

	// List of glob pattern strings. Tags with matching keys will be excluded.
	TagExclude []string `yaml:"tagExclude"`
}

func (cfg Config) Empty() bool {
	return len(cfg.MetricWhitelist) == 0 && len(cfg.MetricBlacklist) == 0 && len(cfg.MetricTagWhitelist) == 0 &&
		len(cfg.MetricTagBlacklist) == 0 && len(cfg.TagInclude) == 0 && len(cfg.TagExclude) == 0
}
