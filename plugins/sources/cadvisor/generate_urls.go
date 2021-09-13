package cadvisor

import (
	"net"
	"net/url"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeLister interface {
	List(opts metav1.ListOptions) (*v1.NodeList, error)
}

const cAdvisorEndpoint = "/metrics/cadvisor"

// GenerateURLs generates cAdvisor prometheus urls to be queried by THIS collector instance
func GenerateURLs(lister NodeLister, myNode string, daemonMode bool, kubeletURL func(ip net.IP, path string) *url.URL) ([]*url.URL, error) {
	nodeList, err := lister.List(metav1.ListOptions{})
	var urls []*url.URL
	if daemonMode {
		for _, node := range nodeList.Items {
			if node.Name == myNode {
				_, ip, _ := util.GetNodeHostnameAndIP(&node)
				urls = append(urls, kubeletURL(ip, cAdvisorEndpoint))
				break
			}
		}
	} else {
		if err != nil {
			return nil, err
		}
		for _, node := range nodeList.Items {
			_, ip, _ := util.GetNodeHostnameAndIP(&node)
			urls = append(urls, kubeletURL(ip, cAdvisorEndpoint))
		}
	}
	return urls, nil
}
