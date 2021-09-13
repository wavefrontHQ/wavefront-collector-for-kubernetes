// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/sources/kubelet/kubelet.go
// Diff against master for changes to the original code.

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

// This file implements a cadvisor datasource, that collects metrics from an instance
// of cadvisor running on a specific host.

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kubelet

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/kubernetes"

	cadvisor "github.com/google/cadvisor/info/v1"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	stats "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

type Host struct {
	IP       net.IP
	Port     int
	Resource string
}

func (h Host) String() string {
	return net.JoinHostPort(h.IP.String(), strconv.Itoa(h.Port))
}

type KubeletClient struct {
	config *KubeletClientConfig
	client *http.Client
}

type ErrNotFound struct {
	endpoint string
}

func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("%q not found", err.endpoint)
}

func sampleContainerStats(stats []*cadvisor.ContainerStats) []*cadvisor.ContainerStats {
	if len(stats) == 0 {
		return []*cadvisor.ContainerStats{}
	}
	return []*cadvisor.ContainerStats{stats[len(stats)-1]}
}

func (kc *KubeletClient) postRequestAndGetValue(client *http.Client, req *http.Request, value interface{}) error {
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body - %v", err)
	}
	if response.StatusCode == http.StatusNotFound {
		return &ErrNotFound{req.URL.String()}
	} else if response.StatusCode == http.StatusForbidden {
		kubernetes.Terminate("Missing ClusterRole resource nodes/stats or nodes/proxy, see https://docs.wavefront.com/kubernetes.html#kubernetes-manual-install")
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed - %q, response: %q", response.Status, string(body))
	}

	kubeletAddr := "[unknown]"
	if req.URL != nil {
		kubeletAddr = req.URL.Host
	}

	log.WithFields(log.Fields{
		"address":  kubeletAddr,
		"response": string(body),
	}).Trace("Raw response from kubelet")

	err = jsoniter.ConfigFastest.Unmarshal(body, value)
	if err != nil {
		return fmt.Errorf("failed to parse output. Response: %q. Error: %v", string(body), err)
	}
	return nil
}

func (kc *KubeletClient) parseStat(containerInfo *cadvisor.ContainerInfo) *cadvisor.ContainerInfo {
	containerInfo.Stats = sampleContainerStats(containerInfo.Stats)
	if len(containerInfo.Aliases) > 0 {
		containerInfo.Name = containerInfo.Aliases[0]
	}
	return containerInfo
}

func (kc *KubeletClient) GetSummary(ip net.IP) (*stats.Summary, error) {
	u := kc.config.BaseURL(ip, "/stats/summary/").String()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	summary := &stats.Summary{}
	client := kc.client
	if client == nil {
		client = http.DefaultClient
	}
	err = kc.postRequestAndGetValue(client, req, summary)
	return summary, err
}

func (kc *KubeletClient) GetPods(ip net.IP) (*v1.PodList, error) {
	u := kc.config.BaseURL(ip, "/pods/").String()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	pods := &v1.PodList{}
	client := kc.client
	if client == nil {
		client = http.DefaultClient
	}
	err = kc.postRequestAndGetValue(client, req, pods)

	return pods, err
}

func (kc *KubeletClient) GetPort() uint {
	return kc.config.Port
}

func NewKubeletClient(kubeletConfig *KubeletClientConfig) (*KubeletClient, error) {
	transport, err := MakeTransport(kubeletConfig)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Transport: transport,
		Timeout:   kubeletConfig.HTTPTimeout,
	}
	return &KubeletClient{
		config: kubeletConfig,
		client: c,
	}, nil
}
