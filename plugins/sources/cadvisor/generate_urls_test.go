package cadvisor

import (
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StubNodeLister v1.NodeList

func (s *StubNodeLister) List(_ metav1.ListOptions) (*v1.NodeList, error) {
	return (*v1.NodeList)(s), nil
}

type ErrorNodeLister string

func (s ErrorNodeLister) List(_ metav1.ListOptions) (*v1.NodeList, error) {
	return nil, errors.New(string(s))
}

func TestGenerateURLs(t *testing.T) {
	nodeLister := &StubNodeLister{Items: []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{{Type: v1.NodeInternalIP, Address: "127.0.0.1"}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{{Type: v1.NodeInternalIP, Address: "127.0.0.2"}},
			},
		},
	}}
	myNode := "node-1"
	kubeletURL := func(ip net.IP, path string) *url.URL {
		return &url.URL{
			Scheme: "https",
			Host:   ip.String() + ":10250",
			Path:   path,
		}
	}

	t.Run("returns an error when it cannot list nodes", func(t *testing.T) {
		expectedErrorStr := "something went wrong"
		_, err := GenerateURLs(ErrorNodeLister(expectedErrorStr), myNode, false, kubeletURL)

		assert.Equal(t, expectedErrorStr, err.Error())
	})

	t.Run("returns an error when a node does not have an IP", func(t *testing.T) {
		_, err := GenerateURLs(
			&StubNodeLister{Items: []v1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}}}},
			myNode,
			false,
			kubeletURL,
		)

		assert.Contains(t, err.Error(), "has no valid hostname and/or IP address")
	})

	t.Run("successfully generates one URL when DaemonMode is true", func(t *testing.T) {
		urls, err := GenerateURLs(nodeLister, myNode, true, kubeletURL)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(urls))
	})

	t.Run("generates urls using the kubeleteURL func", func(t *testing.T) {
		urls, _ := GenerateURLs(nodeLister, myNode, true, kubeletURL)

		assert.Equal(t, urls[0].String(), "https://127.0.0.1:10250/metrics/cadvisor")
	})

	t.Run("successfully generates URLs for each node when DaemonMode is false", func(t *testing.T) {
		urls, err := GenerateURLs(nodeLister, myNode, false, kubeletURL)

		assert.Nil(t, err)
		assert.Equal(t, len(nodeLister.Items), len(urls))
	})
}
