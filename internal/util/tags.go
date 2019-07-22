package util

type GroupedTags struct {
	tags map[string]map[string]string
}

func NewGroupedTags() *GroupedTags {
	return &GroupedTags{
		tags: make(map[string]map[string]string),
	}
}

func (gt *GroupedTags) GetOrAdd(key string) (map[string]string, bool) {
	if val, exists := gt.tags[key]; exists {
		return val, exists
	}
	m := map[string]string{}
	gt.tags[key] = m
	return m, false
}
