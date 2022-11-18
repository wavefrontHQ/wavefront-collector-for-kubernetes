// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	emptyReason       = "they were empty"
	excludeListReason = "they were on an exclude list"
	dedupeReason      = "there were too many tags so we removed tags with duplicate tag values"
	iaasReason        = "there were too many tags so we removed IaaS specific tags"
)

var iaasNameRegex = regexp.MustCompile("^label.*gke*")

// cleanTags removes empty, excluded tags, and tags with duplicate values (if there are too many tags) and returns a map
// that lists removed tag names by their reason for removal
func cleanTags(tags map[string]string, maxCapacity int) map[string][]string {
	removedReasons := map[string][]string{}
	removedReasons[emptyReason] = removeEmptyTags(tags)
	removedReasons[excludeListReason] = excludeTags(tags)
	if len(tags) > maxCapacity {
		removedReasons[dedupeReason] = dedupeTagValues(tags)
	}

	// remove IaaS label tags is over max capacity
	//if len(tags) > maxCapacity {
	//	removedReasons[iaasReason] = removeTagsLabelsMatching(tags, iaasNameRegex, len(tags)-maxCapacity)
	//}
	return removedReasons
}

func logTagCleaningReasons(metricName string, reasons map[string][]string) {
	for reason, tagNames := range reasons {
		if len(tagNames) == 0 {
			continue
		}
		log.Debugf(
			"the following tags were removed from %s because %s: %s",
			metricName, reason, strings.Join(tagNames, ", "),
		)
	}
}

const minDedupeTagValueLen = 10

func dedupeTagValues(tags map[string]string) []string {
	var removedTags []string
	invertedTags := map[string]string{} // tag value -> tag name
	for name, value := range tags {
		if len(value) < minDedupeTagValueLen {
			continue
		}
		if len(invertedTags[value]) == 0 {
			invertedTags[value] = name
		} else if isWinningName(name, invertedTags[value]) {
			removedTags = append(removedTags, invertedTags[value])
			delete(tags, invertedTags[value])
			invertedTags[value] = name
		} else {
			removedTags = append(removedTags, name)
			delete(tags, name)
		}
	}
	return removedTags
}

func isWinningName(name string, prevWinner string) bool {
	return len(name) < len(prevWinner) || (len(name) == len(prevWinner) && name < prevWinner)
}

func removeEmptyTags(tags map[string]string) []string {
	var removed []string
	for name, value := range tags {
		if len(value) == 0 {
			removed = append(removed, name)
			delete(tags, name)
		}
	}
	return removed
}

func removeTagsLabelsMatching(tags map[string]string, regexp *regexp.Regexp, numberToRemove int) []string {
	var removed []string
	count := 0
	tagNames := sortKeys(tags)
	for _, name := range tagNames {
		if regexp.MatchString(name) {
			removed = append(removed, name)
			delete(tags, name)
			count++
			if count >= numberToRemove {
				break
			}
		}
	}
	return removed
}

func excludeTags(tags map[string]string) []string {
	var removed []string
	for name := range tags {
		if excludeTag(name) {
			removed = append(removed, name)
			delete(tags, name)
		}
	}
	return removed
}

func excludeTag(name string) bool {
	for _, excludeName := range excludeTagList {
		if excludeName == name {
			return true
		}
	}
	for _, excludePrefix := range excludeTagPrefixes {
		if strings.HasPrefix(name, excludePrefix) {
			return true
		}
	}
	return false
}
