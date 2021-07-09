// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEncodeTags(t *testing.T) {
	labels := make(map[string]string)
	labels["a"] = "a"
	labels["b"] = "b"

	tags := make(map[string]string)
	EncodeTags(tags, "label.", labels)
	checkTag(tags, "label.a", "a", t)
	checkTag(tags, "label.b", "b", t)
}

func TestEncodePod(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	tags := make(map[string]string)
	EncodeMeta(tags, "pod", pod.ObjectMeta)
	checkTag(tags, "pod", "test", t)
	checkTag(tags, "namespace", "test-ns", t)
}

func TestParam(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{"key1": "value1"},
		},
	}
	p := Param(pod.ObjectMeta, "key1", "cfgValue", "defaultValue")
	if p != "value1" {
		t.Errorf("expected annotation value: %s actual: %s", "value1", p)
	}
	p = Param(pod.ObjectMeta, "key2", "cfgValue", "defaultValue")
	if p != "cfgValue" {
		t.Errorf("expected cfg value: %s actual: %s", "cfgValue", p)
	}
	p = Param(pod.ObjectMeta, "key2", "", "defaultValue")
	if p != "defaultValue" {
		t.Errorf("expected default value: %s actual: %s", "defaultValue", p)
	}
}

func checkTag(tags map[string]string, key, val string, t *testing.T) {
	if len(tags) == 0 {
		t.Error("missing tags")
	}
	if v, ok := tags[key]; ok {
		if v == val {
			return
		}
	}
	t.Errorf("missing tag: %s", key)
}
