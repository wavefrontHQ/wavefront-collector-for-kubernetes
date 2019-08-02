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

package kubelet

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"

	kube_config "github.com/wavefronthq/wavefront-kubernetes-collector/internal/kubernetes"
	kube_client "k8s.io/client-go/rest"
)

const (
	APIVersion = "v1"

	defaultKubeletPort        = 10255
	defaultKubeletHttps       = false
	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
)

func GetKubeConfigs(cfg configuration.SummaySourceConfig) (*kube_client.Config, *KubeletClientConfig, error) {

	kubeConfig, err := kube_config.GetKubeClientConfigFromConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	kubeletPort := defaultKubeletPort
	if len(cfg.KubeletPort) >= 1 {
		kubeletPort, err = strconv.Atoi(cfg.KubeletPort)
		if err != nil {
			return nil, nil, err
		}
	}

	kubeletHttps := defaultKubeletHttps
	if len(cfg.KubeletHttps) >= 1 {
		kubeletHttps, err = strconv.ParseBool(cfg.KubeletHttps)
		if err != nil {
			return nil, nil, err
		}
	}

	log.Infof("Using Kubernetes client with master %q and version %+v\n", kubeConfig.Host, kubeConfig.GroupVersion)
	log.Infof("Using kubelet port %d", kubeletPort)

	kubeletConfig := &KubeletClientConfig{
		Port:            uint(kubeletPort),
		EnableHttps:     kubeletHttps,
		TLSClientConfig: kubeConfig.TLSClientConfig,
		BearerToken:     kubeConfig.BearerToken,
	}
	return kubeConfig, kubeletConfig, nil
}
