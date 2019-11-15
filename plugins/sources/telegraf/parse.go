// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"
)

func initPlugin(input telegraf.Input, conf string) (err error) {
	defer func() {
		// handle panic errors when parsing erroneous configurations
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid telegraf configuration: %v", r)
		}
	}()

	if len(conf) == 0 {
		return fmt.Errorf("missing telegraf configuration")
	}
	err = toml.Unmarshal([]byte(conf), input)
	return
}
