package cadvisor

import (
	"fmt"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeLister interface {
	List(opts metav1.ListOptions) (*v1.NodeList, error)
}

// GenerateURLs generates cAdvisor prometheus urls to be queried by THIS collector instance
func GenerateURLs(lister NodeLister, myNode string, daemonMode bool, baseURL string) ([]string, error) {
	nodeList, err := lister.List(metav1.ListOptions{})
	var urls []string
	if daemonMode {
		for _, node := range nodeList.Items {
			if node.Name == myNode {
				hostname, ip, _ := util.GetNodeHostnameAndIP(&node)
				log.Printf("mynode hostname=%s ip=%s", hostname, ip)
				urls = append(urls, generateCadvisorURL(node.Name, fmt.Sprintf("https://%s:10250", ip)))
				break
			}
		}
	} else {
		if err != nil {
			return nil, err
		}
		for _, node := range nodeList.Items {
			urls = append(urls, generateCadvisorURL(node.Name, baseURL))
		}
	}
	return urls, nil
}

const cadvisorURLPattern = "%s/metrics/cadvisor"

func generateCadvisorURL(nodeName string, baseURL string) string {
	return fmt.Sprintf(cadvisorURLPattern, baseURL)
}
