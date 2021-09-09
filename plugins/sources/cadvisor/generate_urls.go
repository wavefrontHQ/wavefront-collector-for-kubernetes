package cadvisor

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeLister interface {
	List(opts metav1.ListOptions) (*v1.NodeList, error)
}

// GenerateURLs generates prometheus urls to be queried by THIS collector instance for cAdvisor metrics
func GenerateURLs(lister NodeLister, myNode string, daemonMode bool) ([]string, error) {
	var urls []string
	if daemonMode {
		urls = append(urls, generateCadvisorURL(myNode))
	} else {
		nodeList, err := lister.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, node := range nodeList.Items {
			urls = append(urls, generateCadvisorURL(node.Name))
		}
	}
	return urls, nil
}

const cAdvisorURL = "https://kubernetes.default.svc.cluster.local:443/api/v1/nodes/%s/proxy/metrics/cadvisor"

func generateCadvisorURL(nodeName string) string {
	return fmt.Sprintf(cAdvisorURL, nodeName)
}
