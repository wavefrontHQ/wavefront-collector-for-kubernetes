package filter

import "testing"

func TestFromQuery(t *testing.T) {
	vals := make(map[string][]string)
	if FromQuery(vals) != nil {
		t.Errorf("error creating filter")
	}
	vals[MetricWhitelist] = []string{"foo*", "bar*"}
	vals[MetricBlacklist] = []string{"drop*", "etc*"}
	vals[MetricTagWhitelist] = []string{"foo*", "bar*"}
	vals[MetricTagBlacklist] = []string{"drop*", "etc*"}
	vals[TagInclude] = []string{"key1*", "key2*"}
	vals[TagExclude] = []string{"key3*", "key4*"}

	f := FromQuery(vals)
	if f == nil {
		t.Errorf("error creating filter")
	}
	gf, ok := f.(*globFilter)
	if !ok {
		t.Errorf("error creating filter")
	}

	if gf.namePass == nil || gf.nameDrop == nil || gf.tagPass == nil || gf.tagDrop == nil ||
		gf.tagInclude == nil || gf.tagExclude == nil {
		t.Errorf("error creating filter")
	}
}

func TestParseValue(t *testing.T) {
	slice, err := parseValue("[foo*,bar*]")
	if err != nil {
		t.Errorf("error parsing value: %q", err)
	}
	expected := []string{"foo*", "bar*"}
	if !equalSlice(slice, expected) {
		t.Errorf("error parsing value, expected:%s actual:%s", expected, slice)
	}
}

func TestParseFilters(t *testing.T) {
	actual := parseFilters([]string{"env:[dev*,staging*]", "cluster:[*dev*,*staging*]"})
	expected := map[string][]string{
		"env":     {"dev*", "staging*"},
		"cluster": {"*dev*", "*staging*"},
	}
	if !equalMap(actual, expected) {
		t.Errorf("error parsing filters, expected: %q actual: %q", expected, actual)
	}
}

func equalMap(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if _, ok := b[k]; !ok {
			return false
		}
		if !equalSlice(v, b[k]) {
			return false
		}
	}
	return true
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
