// Based on https://github.com/kubernetes-retired/heapster/blob/master/common/kubernetes/configs.go
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

package kubernetes

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_rest "k8s.io/client-go/rest"
	kubeClientCmd "k8s.io/client-go/tools/clientcmd"
	kubeClientCmdApi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	APIVersion        = "v1"
	DefaultAPIService = "kubernetes.default.svc.cluster.local"

	defaultKubeletPort        = 10255
	defaultKubeletHttps       = false
	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
)

func getConfigOverrides(uri *url.URL) (*kubeClientCmd.ConfigOverrides, error) {
	kubeConfigOverride := kubeClientCmd.ConfigOverrides{
		ClusterInfo: kubeClientCmdApi.Cluster{},
	}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}

	opts := uri.Query()

	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}

	return &kubeConfigOverride, nil
}

func getConfigOverridesFromConfig(cfg configuration.SummaySourceConfig) (*kubeClientCmd.ConfigOverrides, error) {
	//TODO: validate this returns the same values as with flags
	kubeConfigOverride := kubeClientCmd.ConfigOverrides{
		ClusterInfo: kubeClientCmdApi.Cluster{},
	}

	if cfg.URL != "" {
		uri, err := url.Parse(cfg.URL)
		if err != nil {
			return nil, err
		}
		if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
			kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
		}
	}

	if len(cfg.Insecure) > 0 {
		insecure, err := strconv.ParseBool(cfg.Insecure)
		if err != nil {
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}
	return &kubeConfigOverride, nil
}

func GetKubeClientConfigFromConfig(cfg configuration.SummaySourceConfig) (*kube_rest.Config, error) {
	var (
		kubeConfig *kube_rest.Config
		err        error
	)

	configOverrides, err := getConfigOverridesFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	inClusterConfig := defaultInClusterConfig
	if len(cfg.InClusterConfig) > 0 {
		inClusterConfig, err = strconv.ParseBool(cfg.InClusterConfig)
		if err != nil {
			return nil, err
		}
	}

	if inClusterConfig {
		kubeConfig, err = kube_rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		if configOverrides.ClusterInfo.Server != "" {
			kubeConfig.Host = configOverrides.ClusterInfo.Server
		}
		kubeConfig.GroupVersion = &schema.GroupVersion{Version: APIVersion}
		kubeConfig.Insecure = configOverrides.ClusterInfo.InsecureSkipTLSVerify
		if configOverrides.ClusterInfo.InsecureSkipTLSVerify {
			kubeConfig.TLSClientConfig.CAFile = ""
		}
	} else {
		authFile := ""
		if len(cfg.Auth) > 0 {
			authFile = cfg.Auth
		}

		if authFile != "" {
			// Load structured kubeconfig data from the given path.
			loader := &kubeClientCmd.ClientConfigLoadingRules{ExplicitPath: authFile}
			loadedConfig, err := loader.Load()
			if err != nil {
				return nil, err
			}

			// Flatten the loaded data to a particular restclient.Config based on the current context.
			if kubeConfig, err = kubeClientCmd.NewNonInteractiveClientConfig(
				*loadedConfig,
				loadedConfig.CurrentContext,
				&kubeClientCmd.ConfigOverrides{},
				loader).ClientConfig(); err != nil {
				return nil, err
			}
		} else {
			kubeConfig = &kube_rest.Config{
				Host: configOverrides.ClusterInfo.Server,
				TLSClientConfig: kube_rest.TLSClientConfig{
					Insecure: configOverrides.ClusterInfo.InsecureSkipTLSVerify,
				},
			}
			kubeConfig.GroupVersion = &schema.GroupVersion{Version: APIVersion}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes master url specified")
	}

	useServiceAccount := defaultUseServiceAccount
	if len(cfg.UseServiceAccount) >= 1 {
		useServiceAccount, err = strconv.ParseBool(cfg.UseServiceAccount)
		if err != nil {
			return nil, err
		}
	}

	if useServiceAccount {
		// If a readable service account token exists, then use it
		if contents, err := ioutil.ReadFile(defaultServiceAccountFile); err == nil {
			kubeConfig.BearerToken = string(contents)
		}
	}
	kubeConfig.ContentType = "application/vnd.kubernetes.protobuf"

	return kubeConfig, nil
}
