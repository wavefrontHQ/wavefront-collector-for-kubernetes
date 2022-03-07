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

// Copyright 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
)

// ResourceKey is used to group metric sets by a specific k8s resource.
// A ResourceKey should be treated as an opaque identifier.
// A ResourceKey should not be serialized as its format is not public.
type ResourceKey string

func (s ResourceKey) String() string {
	return string(s)
}

func (s ResourceKey) Append(t ResourceKey) ResourceKey {
	return ResourceKey(s.String() + t.String())
}

func PodContainerKey(namespace, podName, containerName string) ResourceKey {
	return ResourceKey(fmt.Sprintf("namespace:%s/pod:%s/container:%s", namespace, podName, containerName))
}

func PodKey(namespace, podName string) ResourceKey {
	return ResourceKey(fmt.Sprintf("namespace:%s/pod:%s", namespace, podName))
}

func NamespaceKey(namespace string) ResourceKey {
	return ResourceKey(fmt.Sprintf("namespace:%s", namespace))
}

func NodeKey(node string) ResourceKey {
	return ResourceKey(fmt.Sprintf("node:%s", node))
}

func NodeContainerKey(node, container string) ResourceKey {
	return ResourceKey(fmt.Sprintf("node:%s/container:%s", node, container))
}

func ClusterKey() ResourceKey {
	return "cluster"
}

type ResourceKeys []ResourceKey

func (s ResourceKeys) Len() int {
	return len(s)
}

func (s ResourceKeys) Less(i, j int) bool {
	return string(s[i]) < string(s[j])
}

func (s ResourceKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
