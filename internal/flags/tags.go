package flags

import (
	"github.com/golang/glog"
	"strings"
)

// Decodes tags of the form "tag=key:value"
func DecodeTags(vals map[string][]string) map[string]string {
	if vals == nil {
		return nil
	}
	var tags map[string]string
	if len(vals["tag"]) > 0 {
		tags = make(map[string]string)
		tagList := vals["tag"]
		for _, tag := range tagList {
			s := strings.Split(tag, ":")
			if len(s) == 2 {
				k, v := s[0], s[1]
				tags[k] = v
			} else {
				glog.Warning("invalid tag ", tag)
			}
		}
	}
	return tags
}
