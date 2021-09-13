package cadvisor

import (
	"errors"
	"testing"

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
	t.Fatal("TODO: Re-implement tests")
	// nodeLister := &StubNodeLister{Items: []v1.Node{
	// 	{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
	// 	{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
	// }}
	// myNode := "node-1"
	//
	// t.Run("when DaemonMode is true", func(t *testing.T) {
	// 	t.Run("successfully generates one URL", func(t *testing.T) {
	// 		configs, err := GenerateURLs(nodeLister, myNode, true, "")
	//
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, 1, len(configs))
	// 	})
	//
	// 	t.Run("the url contains myNode", func(t *testing.T) {
	// 		urls, _ := GenerateURLs(nodeLister, myNode, true, "myHost")
	//
	// 		assert.Contains(t, urls[0], myNode)
	// 		assert.Contains(t, urls[0], "myHost")
	// 	})
	// })
	//
	// t.Run("when DaemonMode is false", func(t *testing.T) {
	// 	t.Run("successfully produces URLs foreach node", func(t *testing.T) {
	// 		// This is the case when the leader wants to query all nodes instead of having each node's collector do it
	// 		configs, err := GenerateURLs(nodeLister, myNode, false, "")
	//
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, len(nodeLister.Items), len(configs))
	// 	})
	//
	// 	t.Run("interpolates each node name into a URL", func(t *testing.T) {
	// 		urls, _ := GenerateURLs(nodeLister, myNode, false, "myHost")
	//
	// 		for i, node := range nodeLister.Items {
	// 			assert.Contains(t, urls[i], node.Name)
	// 			assert.Contains(t, urls[i], "myHost")
	// 		}
	// 	})
	//
	// 	t.Run("returns an error when it cannot list nodes", func(t *testing.T) {
	// 		expectedErrorStr := "something went wrong"
	// 		_, err := GenerateURLs(ErrorNodeLister(expectedErrorStr), myNode, false, "")
	//
	// 		assert.Equal(t, expectedErrorStr, err.Error())
	// 	})
	// })
}
