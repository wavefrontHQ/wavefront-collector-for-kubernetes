package prometheus

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var matchNSName = regexp.MustCompile("^([^\\.]+).([^\\.]+)\\.svc")

type Instance struct {
	Host string
	Tags map[string]string
}

type LookupInstances func(host string) ([]Instance, error)

func InstancesFromEndpoints(client corev1.EndpointsGetter) LookupInstances {
	return func(host string) ([]Instance, error) {
		matches := matchNSName.FindStringSubmatch(host)
		if len(matches) == 0 {
			return nil, errors.New("host is not a kubernetes service")
		}
		endpoints, err := client.Endpoints(matches[2]).Get(context.Background(), matches[1], metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return instancesFromEndpoints(endpoints)
	}
}

func instancesFromEndpoints(endpoints *apicorev1.Endpoints) ([]Instance, error) {
	var instances []Instance
	for _, subset := range endpoints.Subsets {
		toAdd, err := instancesFromSubset(subset)
		if err != nil {
			return nil, err
		}
		instances = append(instances, toAdd...)
	}
	return removeDuplicateInstances(instances), nil
}

func instancesFromSubset(subset apicorev1.EndpointSubset) ([]Instance, error) {
	port := choosePort(subset.Ports)
	if port == -1 {
		return nil, errors.New("could not find either HTTP or HTTPS port")
	}
	instances := make([]Instance, 0, len(subset.Addresses))
	for _, address := range subset.Addresses {
		host := fmt.Sprintf("%s:%d", address.IP, port)
		instances = append(instances, Instance{
			Host: host,
			Tags: map[string]string{"instance": host},
		})
	}
	return instances, nil
}

func removeDuplicateInstances(instances []Instance) []Instance {
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Host < instances[j].Host
	})
	var unique []Instance
	for _, instance := range instances {
		if len(unique) == 0 || unique[len(unique)-1].Host != instance.Host {
			unique = append(unique, instance)
		}
	}
	return unique
}

func choosePort(ports []apicorev1.EndpointPort) int32 {
	httpPort := int32(-1)
	for _, port := range ports {
		if port.Name == "http" {
			httpPort = port.Port
		}
		if port.Name == "https" {
			return port.Port
		}
	}
	return httpPort
}

func InstanceFromHost(instance string) ([]Instance, error) {
	return []Instance{{instance, nil}}, nil
}
