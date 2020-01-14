// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"
	kubeFlag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
)

var (
	version string
	commit  string
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	opt := options.NewCollectorRunOptions()
	opt.AddFlags(pflag.CommandLine)
	kubeFlag.InitFlags()

	if opt.Version {
		fmt.Println(fmt.Sprintf("version: %s\ncommit: %s", version, commit))
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	waitForStop()
}

func waitForStop() {
	select {}
}
