// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//+build !race

package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	counter := 0
	retryFunc := func() { counter++ }

	t.Run("Test stop", func(t *testing.T) {
		stopCh := make(chan struct{})
		go Retry(retryFunc, 1*time.Second, stopCh)
		time.Sleep(1500 * time.Millisecond)
		close(stopCh)
		assert.Equal(t, 1, counter)
	})

	t.Run("Test delay before first retry", func(t *testing.T) {
		counter = 0
		stopCh := make(chan struct{})
		go Retry(retryFunc, 1*time.Second, stopCh)
		time.Sleep(500 * time.Millisecond)
		close(stopCh)
		assert.Equal(t, 0, counter)
	})
}
