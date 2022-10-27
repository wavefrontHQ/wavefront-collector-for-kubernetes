package prometheus

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func LookupByEndpoints(getEndpoints func() (*corev1.Endpoints, error)) LookupHost {
	return func(host string) (addrs []string, err error) {
		endpoints, err := getEndpoints()
		if err != nil {
			return nil, err
		}
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				for _, port := range subset.Ports {
					if (port.Name == "http" || port.Name == "https") && port.Protocol == "TCP" {
						addrs = append(addrs, fmt.Sprintf("%s:%d", address.IP, port.Port))
					}
				}
			}
		}
		return addrs, nil
	}
}
