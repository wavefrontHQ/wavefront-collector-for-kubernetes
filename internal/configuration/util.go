// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configuration

import "time"

func GetStringValue(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func GetDurationValue(value, defaultValue time.Duration) time.Duration {
	if value != 0 {
		return value
	}
	return defaultValue
}
