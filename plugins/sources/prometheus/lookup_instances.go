package prometheus

import (
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var matchNSName = regexp.MustCompile("^([^\\.]+).([^\\.]+).svc")

func LookupByEndpoints(client corev1.CoreV1Interface) LookupInstances {
	return func(host string) (instances []Instance, err error) {
		matches := matchNSName.FindStringSubmatch(host)
		if len(matches) == 0 {
			return nil, errors.New("does not match expected hostname format")
		}
		endpoints, err := client.Endpoints(matches[2]).Get(context.Background(), matches[1], metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				for _, port := range subset.Ports {
					if (port.Name == "http" || port.Name == "https") && port.Protocol == "TCP" {
						instanceAddress := fmt.Sprintf("%s:%d", address.IP, port.Port)
						instances = append(instances, Instance{
							Address: instanceAddress,
							Tags:    map[string]string{"instance": instanceAddress},
						})
					}
				}
			}
		}
		return instances, nil
	}
}

func NoopLookupInstances(instance string) ([]Instance, error) {
	return []Instance{{instance, nil}}, nil
}

type Instance struct {
	Address string
	Tags    map[string]string
}

type LookupInstances func(host string) ([]Instance, error)
