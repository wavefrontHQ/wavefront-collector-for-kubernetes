// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EncodeTags(destTags map[string]string, prefix string, tags map[string]string) {
	if len(tags) == 0 {
		return
	}
	for k, v := range tags {
		if k != "pod-template-hash" && len(k) > 0 && len(v) > 0 {
			key := fmt.Sprintf("%s%s", prefix, k)
			destTags[key] = v
		}
	}
}

func EncodeMeta(tags map[string]string, kind string, meta metav1.ObjectMeta) {
	tags[kind] = meta.Name
	if meta.Namespace != "" {
		tags["namespace"] = meta.Namespace
	}
}

func Param(meta metav1.ObjectMeta, annotation, cfgVal, defaultVal string) string {
	value := ""
	// give precedence to annotation
	if annotation != "" {
		value = meta.GetAnnotations()[annotation]
	}
	if value == "" {
		// then config
		value = cfgVal
	}
	if value == "" {
		// then default
		value = defaultVal
	}
	return value
}
