// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package flags

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
				log.Warning("invalid tag ", tag)
			}
		}
	}
	return tags
}

func DecodeValue(vals map[string][]string, name string) string {
	value := ""
	if len(vals[name]) > 0 {
		value = vals[name][0]
	}
	return value
}

func DecodeBoolean(vals map[string][]string, name string) bool {
	value := false
	if len(vals[name]) > 0 {
		var err error
		value, err = strconv.ParseBool(vals[name][0])
		if err != nil {
			return false
		}
	}
	return value
}

func ParseDuration(vals url.Values, prop string, def time.Duration) time.Duration {
	if len(vals[prop]) > 0 {
		res, err := time.ParseDuration(vals[prop][0])
		if err != nil {
			log.Errorf("error parsing '%s' propertie: %v", prop, err)
		} else {
			return res
		}
	}
	return def
}
