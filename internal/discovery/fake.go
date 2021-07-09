// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FakeService(name, namespace, ip string) *v1.Service {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: ip,
		},
	}
	return &service
}

func FakePod(name, namespace, ip string) *v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1.PodStatus{
			PodIP: ip,
		},
	}
	return &pod
}

type FakeDiscoverer struct{}

func (f *FakeDiscoverer) Discover(resource Resource) {}
func (f *FakeDiscoverer) Delete(resource Resource)   {}
