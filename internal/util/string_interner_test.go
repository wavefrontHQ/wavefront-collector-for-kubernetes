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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSelf(t *testing.T) {
	interner := NewStringInterner()

	string1 := "foo"
	string2 := interner.Intern(string1)
	assert.Equal(t, "foo", *string2)
}

func TestSingleReference(t *testing.T) {
	interner := NewStringInterner()

	firstResult := interner.Intern("foo")
	assert.Equal(t, "foo", *firstResult)

	secondResult := interner.Intern("foo")
	assert.True(t, firstResult == secondResult)
}

func TestGetDifferentReference(t *testing.T) {
	interner := NewStringInterner()

	firstString := interner.Intern("foo")

	differentString := interner.Intern("bar")
	assert.Equal(t, "bar", *differentString)
	assert.False(t, firstString == differentString)
}
