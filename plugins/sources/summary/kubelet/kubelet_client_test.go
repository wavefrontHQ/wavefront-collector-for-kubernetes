// Copyright 2014 Google Inc. All Rights Reserved.
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

package kubelet

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/kubernetes"
	kube_api "k8s.io/api/core/v1"
	util "k8s.io/client-go/util/testing"
)

func TestGetPods(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		content, err := ioutil.ReadFile("k8s_api_pods.json")
		require.NoError(t, err)

		handler := util.FakeHandler{
			StatusCode:   200,
			RequestBody:  "",
			ResponseBody: string(content),
			T:            t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()
		kubeletClient := KubeletClient{}
		u, _ := url.Parse(server.URL)
		_, port, _ := net.SplitHostPort(u.Host)
		portInt, _ := strconv.Atoi(port)
		pods, err := kubeletClient.GetPods(Host{
			IP:       net.ParseIP("127.0.0.1"),
			Port:     portInt,
			Resource: "",
		})

		require.NoError(t, err)
		require.Len(t, pods.Items, 7)
		assert.Equal(t, pods.Items[0].Status.Phase, kube_api.PodSucceeded, "Expected Pod status phase to be succeeded")
		assert.Equal(t, pods.Items[5].Status.Phase, kube_api.PodFailed, "Expected Pod status phase to be failed")

	})

	t.Run("forbidden", func(t *testing.T) {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
		kubernetes.UseTerminateTestMode()

		server := httptest.NewServer(handlerFunc)
		defer server.Close()

		kubeletClient := KubeletClient{}

		u, _ := url.Parse(server.URL)
		_, port, _ := net.SplitHostPort(u.Host)
		portInt, _ := strconv.Atoi(port)
		kubeletClient.GetPods(Host{
			IP:       net.ParseIP("127.0.0.1"),
			Port:     portInt,
			Resource: "",
		})

		assert.Equal(t, "Missing ClusterRole resource nodes/stats or nodes/proxy, see https://docs.wavefront.com/kubernetes.html#kubernetes-manual-install", kubernetes.TerminationMessage)
	})

}
