// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func TestDefault(t *testing.T) {
	actual := initializeAndCopy(t,
		",",
		[]string{},
		[]string{})

	expected := map[string]string{
		metrics.LabelLabels.Key: "colour:red," + metrics.LabelLabels.Key + ":preorder;configurable,name:bike,price:too_high,weight:10kg",
		"somelabel":             "somevalue",
	}
	assert.Equal(t, expected, actual)
}

func TestSeparator(t *testing.T) {
	actual := initializeAndCopy(t,
		"-",
		[]string{},
		[]string{})

	expected := map[string]string{
		metrics.LabelLabels.Key: "colour:red-" + metrics.LabelLabels.Key + ":preorder;configurable-name:bike-price:too_high-weight:10kg",
		"somelabel":             "somevalue",
	}
	assert.Equal(t, expected, actual)
}

func TestStoredLabels(t *testing.T) {
	actual := initializeAndCopy(t,
		",",
		[]string{"name", "price", "copiedlabels=" + metrics.LabelLabels.Key, "unknown"},
		[]string{})

	expected := map[string]string{
		"name":                  "bike",
		"price":                 "too_high",
		"copiedlabels":          "preorder;configurable",
		metrics.LabelLabels.Key: "colour:red," + metrics.LabelLabels.Key + ":preorder;configurable,name:bike,price:too_high,weight:10kg",
		"somelabel":             "somevalue",
	}
	assert.Equal(t, expected, actual)
}

func TestIgnoredLabels(t *testing.T) {
	actual := initializeAndCopy(t,
		",",
		[]string{},
		[]string{"colour", "weight", "unknown"})

	expected := map[string]string{
		metrics.LabelLabels.Key: metrics.LabelLabels.Key + ":preorder;configurable,name:bike,price:too_high",
		"somelabel":             "somevalue",
	}
	assert.Equal(t, expected, actual)
}

func TestAll(t *testing.T) {
	actual := initializeAndCopy(t,
		"-",
		[]string{"name", "colour", "copiedlabels=" + metrics.LabelLabels.Key, "price", "weight", "unknown"},
		[]string{"colour", metrics.LabelLabels.Key, "price", "unknown"})

	expected := map[string]string{
		"name":                  "bike",
		"colour":                "red",
		"price":                 "too_high",
		"weight":                "10kg",
		"copiedlabels":          "preorder;configurable",
		metrics.LabelLabels.Key: "name:bike-weight:10kg",
		"somelabel":             "somevalue",
	}
	assert.Equal(t, expected, actual)
}

func initializeAndCopy(t *testing.T, separator string, storedLabels []string, ignoredLabels []string) map[string]string {
	lc, err := NewLabelCopier(separator, storedLabels, ignoredLabels)
	if err != nil {
		t.Fatalf("Could not create LabelCopier: %v", err)
	}

	labels := map[string]string{
		"name":                  "bike",
		"colour":                "red",
		"price":                 "too_high",
		"weight":                "10kg",
		metrics.LabelLabels.Key: "preorder;configurable",
	}

	out := map[string]string{
		"somelabel": "somevalue",
	}

	lc.Copy(labels, out)
	return out
}
