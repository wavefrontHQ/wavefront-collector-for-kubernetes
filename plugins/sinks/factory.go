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

package sinks

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks/wavefront"
)

type SinkFactory struct {
}

func (this *SinkFactory) Build(uri flags.Uri) (metrics.DataSink, error) {
	switch uri.Key {
	case "wavefront":
		return wavefront.NewWavefrontSink(&uri.Val)
	default:
		return nil, fmt.Errorf("sink not recognized: %s", uri.Key)
	}
}

func (this *SinkFactory) BuildAll(uris flags.Uris) []metrics.DataSink {
	result := make([]metrics.DataSink, 0, len(uris))

	for _, uri := range uris {
		sink, err := this.Build(uri)
		if err != nil {
			log.Errorf("Failed to create %v sink: %v", uri, err)
			continue
		}
		result = append(result, sink)
	}

	if len([]flags.Uri(uris)) != 0 && len(result) == 0 {
		log.Fatal("No available sink to use")
	}
	return result
}

func NewSinkFactory() *SinkFactory {
	return &SinkFactory{}
}
