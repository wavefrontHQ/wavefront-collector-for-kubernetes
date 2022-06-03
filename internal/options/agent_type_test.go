package options

import (
    "k8s.io/apimachinery/pkg/fields"
    "testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentType(t *testing.T) {
	t.Run("ScrapeCluster", func(t *testing.T) {
		assert.True(t, AllAgentType.ScrapeCluster())
		assert.True(t, LegacyAgentType.ScrapeCluster())
		assert.False(t, NodeAgentType.ScrapeCluster())
		assert.True(t, ClusterAgentType.ScrapeCluster())
	})

	t.Run("ScrapeNodes", func(t *testing.T) {
		assert.Equal(t, "all", AllAgentType.ScrapeNodes())
		assert.Equal(t, "own", LegacyAgentType.ScrapeNodes())
		assert.Equal(t, "own", NodeAgentType.ScrapeNodes())
		assert.Equal(t, "none", ClusterAgentType.ScrapeNodes())
	})

    t.Run("Pod field selector", func(t *testing.T) {
        assert.Equal(t, fields.Everything(), AllAgentType.PodFieldSelector("nodeName"))
        assert.Equal(t, "spec.nodeName=nodeName", LegacyAgentType.PodFieldSelector("nodeName").String())
        assert.Equal(t, "spec.nodeName=nodeName", NodeAgentType.PodFieldSelector("nodeName").String())
        assert.Equal(t, fields.Nothing(), ClusterAgentType.PodFieldSelector("nodeName"))
    })

    t.Run("Node field selector", func(t *testing.T) {
        assert.Equal(t, fields.Everything(), AllAgentType.NodeFieldSelector("nodeName"))
        assert.Equal(t, "metadata.name=nodeName", LegacyAgentType.NodeFieldSelector("nodeName").String())
        assert.Equal(t, "metadata.name=nodeName", NodeAgentType.NodeFieldSelector("nodeName").String())
        assert.Equal(t, fields.Nothing(), ClusterAgentType.NodeFieldSelector("nodeName"))
    })

    t.Run("Should scrape nodes", func(t *testing.T) {
        assert.True(t, AllAgentType.ShouldScrapeAnyNodes())
        assert.True(t, LegacyAgentType.ShouldScrapeAnyNodes())
        assert.True(t, NodeAgentType.ShouldScrapeAnyNodes())
        assert.False(t, ClusterAgentType.ShouldScrapeAnyNodes())
    })
}
