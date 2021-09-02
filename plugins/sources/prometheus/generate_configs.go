package prometheus

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeLister interface {
	List(opts metav1.ListOptions) (*v1.NodeList, error)
}

// GenerateConfigs generates prometheus configs based on the presence of the template variable {{.NodeName}} in the URL
func GenerateConfigs(config configuration.PrometheusSourceConfig, lister NodeLister, getMyNode func() string) ([]configuration.PrometheusSourceConfig, error) {
	if strings.Contains(config.URL, "{{.NodeName}}") {
		return interpolateVariables(config, lister, getMyNode)
	} else {
		return []configuration.PrometheusSourceConfig{config}, nil
	}
}

func interpolateVariables(config configuration.PrometheusSourceConfig, lister NodeLister, getMyNode func() string) ([]configuration.PrometheusSourceConfig, error) {
	var configs []configuration.PrometheusSourceConfig
	nodeNames, err := getNodeNamesForConfig(config, lister, getMyNode)
	if err != nil {
		return nil, err
	}
	tempURL, buffer, err := prepareTemplate(config)
	if err != nil {
		return nil, err
	}
	for _, nodeName := range nodeNames {
		err = executeTemplate(buffer, tempURL, nodeName)
		if err != nil {
			return nil, err
		}
		newConfig := config
		newConfig.URL = buffer.String()
		configs = append(configs, newConfig)
	}
	return configs, nil
}

func getNodeNamesForConfig(config configuration.PrometheusSourceConfig, lister NodeLister, getMyNode func() string) ([]string, error) {
	var nodeNames []string
	if config.PerCluster {
		nodeList, err := lister.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, node := range nodeList.Items {
			nodeNames = append(nodeNames, node.Name)
		}
	} else {
		nodeNames = append(nodeNames, getMyNode())
	}
	return nodeNames, nil
}

func prepareTemplate(config configuration.PrometheusSourceConfig) (*template.Template, *bytes.Buffer, error) {
	tempURL, err := template.New("").Parse(config.URL)
	if err != nil {
		return nil, nil, err
	}
	buffer := bytes.NewBuffer(nil)
	return tempURL, buffer, nil
}

type urlTemplateEnv struct {
	NodeName string
}

func executeTemplate(buffer *bytes.Buffer, tempURL *template.Template, nodeName string) error {
	buffer.Reset()
	return tempURL.Execute(buffer, urlTemplateEnv{NodeName: nodeName})
}
