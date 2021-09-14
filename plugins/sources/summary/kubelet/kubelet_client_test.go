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
		ip, port, close := setupTestServer(t, http.StatusOK)
		defer close()

		kubeletClient := KubeletClient{config: &KubeletClientConfig{Port: port}}
		pods, err := kubeletClient.GetPods(ip)

		require.NoError(t, err)
		require.Len(t, pods.Items, 7)
		assert.Equal(t, pods.Items[0].Status.Phase, kube_api.PodSucceeded, "Expected Pod status phase to be succeeded")
		assert.Equal(t, pods.Items[5].Status.Phase, kube_api.PodFailed, "Expected Pod status phase to be failed")

	})

	t.Run("forbidden", func(t *testing.T) {
		kubernetes.UseTerminateTestMode()
		ip, port, close := setupTestServer(t, http.StatusForbidden)
		defer close()

		kubeletClient := KubeletClient{config: &KubeletClientConfig{Port: port}}
		kubeletClient.GetPods(ip)

		assert.Equal(t, "Missing ClusterRole resource nodes/stats or nodes/proxy, see https://docs.wavefront.com/kubernetes.html#kubernetes-manual-install", kubernetes.TerminationMessage)
	})

}

func setupTestServer(t *testing.T, status int) (net.IP, uint, func()) {
	content, err := ioutil.ReadFile("k8s_api_pods.json")
	require.NoError(t, err)
	handler := util.FakeHandler{
		StatusCode:   status,
		RequestBody:  "",
		ResponseBody: string(content),
		T:            t,
	}
	server := httptest.NewServer(&handler)
	mockServerUrl, _ := url.Parse(server.URL)
	_, port, _ := net.SplitHostPort(mockServerUrl.Host)
	mockPort, _ := strconv.ParseUint(port, 10, 64)
	return net.ParseIP("127.0.0.1"), uint(mockPort), server.Close
}
