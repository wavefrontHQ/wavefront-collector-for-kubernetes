package reporting

import (
	"net/url"
	"sort"
	"strings"
)

// EncodeKey encodes the metric name and tags into a unique key.
func EncodeKey(key string, tags map[string]string) string {
	if len(tags) == 0 {
		return url.QueryEscape(key)
	}

	//sort the tags to ensure the key is always the same when getting or setting
	sortedKeys := make([]string, len(tags))
	i := 0
	for k := range tags {
		sortedKeys[i] = k
		i++
	}
	sort.Strings(sortedKeys)
	keyAppend := url.QueryEscape(key) + "["
	for i := range sortedKeys {
		keyAppend += url.QueryEscape(sortedKeys[i]) + "=" + url.QueryEscape(tags[sortedKeys[i]]) + "&"
	}
	keyAppend = strings.TrimSuffix(keyAppend, "&")
	keyAppend += "]"
	return keyAppend
}

// DecodeKey decodes a metric key into a metric name and tag string
func DecodeKey(key string) (string, map[string]string) {
	if strings.Contains(key, "[") == false {
		name, _ := url.QueryUnescape(key)
		return name, map[string]string{}
	}

	parts := strings.Split(key, "[")
	name, _ := url.QueryUnescape(parts[0])
	tagStr := parts[1]
	tagStr = tagStr[0 : len(tagStr)-1]

	tags := strings.Split(tagStr, "&")
	tagsMap := make(map[string]string)
	for _, pair := range tags {
		z := strings.Split(pair, "=")
		k, _ := url.QueryUnescape(z[0])
		v, _ := url.QueryUnescape(z[1])
		tagsMap[k] = v
	}

	return name, tagsMap
}

func hostTagString(hostTags map[string]string) string {
	htStr := ""
	for k, v := range hostTags {
		htStr += " " + k + "=\"" + v + "\""
	}
	return htStr
}
