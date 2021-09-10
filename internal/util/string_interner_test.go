// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
