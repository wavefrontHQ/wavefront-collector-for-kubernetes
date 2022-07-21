// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build !race
// +build !race

package experimental

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFeatureFlags(t *testing.T) {
	assert.False(t, IsEnabled(ClusterSource), "Error :: Feature cluster-scope is expected to be disabled.")
	EnableFeature(ClusterSource)
	assert.True(t, IsEnabled(ClusterSource), "Error :: Feature cluster-scope is expected to be enabled.")
	DisableFeature(ClusterSource)
	assert.False(t, IsEnabled(ClusterSource), "Error :: Feature cluster-scope is expected to be disabled.")
}
