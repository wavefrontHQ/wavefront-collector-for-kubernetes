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
    "net/http/httptest"
    "net/url"
    "strconv"
    "testing"

    "github.com/stretchr/testify/require"
    util "k8s.io/client-go/util/testing"
)

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
