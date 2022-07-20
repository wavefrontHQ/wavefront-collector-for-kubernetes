// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build !race
// +build !race

package experimental

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestIsEnabled(t *testing.T) {
	t.Run("Test feature enabled", func(t *testing.T) {
        EnableFeature("cluster-scope")
		assert.True(t, IsEnabled("cluster-scope"), "Error :: Feature cluster-scope is expected to be enabled.")
	})
}
