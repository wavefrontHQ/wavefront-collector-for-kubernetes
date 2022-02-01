package cadvisor

import (
	"context"
	"net"
	"net/url"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeLister interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error)
}

const cAdvisorEndpoint = "/metrics/cadvisor"

// GenerateURLs generates cAdvisor prometheus urls to be queried by THIS collector instance
func GenerateURLs(lister NodeLister, myNode string, daemonMode bool, kubeletURL func(ip net.IP, path string) *url.URL) ([]*url.URL, error) {
	nodeList, err := lister.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var urls []*url.URL
	for _, node := range nodeList.Items {
		_, ip, err := util.GetNodeHostnameAndIP(&node)
		if err != nil {
			return nil, err
		}
		kubeletURL := kubeletURL(ip, cAdvisorEndpoint)
		if daemonMode {
			if node.Name == myNode {
				urls = append(urls, kubeletURL)
				break
			}
		} else {
			urls = append(urls, kubeletURL)
		}
	}
	return urls, nil
}
