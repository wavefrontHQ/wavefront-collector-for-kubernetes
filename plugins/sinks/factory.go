// Copyright 2015 Google Inc. All Rights Reserved.
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

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sinks

import (
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks/wavefront"
)

type SinkFactory struct {
}

func (factory *SinkFactory) Build(cfg configuration.WavefrontSinkConfig) (wavefront.WavefrontSink, error) {
	return wavefront.NewWavefrontSink(cfg)
}

func (factory *SinkFactory) BuildAll(cfgs []*configuration.WavefrontSinkConfig) []wavefront.WavefrontSink {
	result := make([]wavefront.WavefrontSink, 0, len(cfgs))

	for _, cfg := range cfgs {
		sink, err := factory.Build(*cfg)
		if err != nil {
			log.Errorf("Failed to create sink: %v", err)
			continue
		}
		result = append(result, sink)
	}

	if len(result) == 0 {
		log.Fatal("No available sink to use")
	}
	return result
}

func NewSinkFactory() *SinkFactory {
	return &SinkFactory{}
}
