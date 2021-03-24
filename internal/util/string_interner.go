// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

type StringInterner map[string]*string

func NewStringInterner() StringInterner {
	return make(map[string]*string)
}

func (interner StringInterner) Intern(s string) *string {
	if interned, ok := interner[s]; ok {
		return interned
	}
	interner[s] = &s
	return &s
}
