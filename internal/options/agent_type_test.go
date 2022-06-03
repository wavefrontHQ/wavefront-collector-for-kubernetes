package options

import (
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

	t.Run("ScrapeAnyNodes", func(t *testing.T) {
		assert.True(t, AllAgentType.ScrapeAnyNodes())
		assert.True(t, LegacyAgentType.ScrapeAnyNodes())
		assert.True(t, NodeAgentType.ScrapeAnyNodes())
		assert.False(t, ClusterAgentType.ScrapeAnyNodes())
	})

	t.Run("ScrapeOnlyOwnNode", func(t *testing.T) {
		assert.False(t, AllAgentType.ScrapeOnlyOwnNode())
		assert.True(t, LegacyAgentType.ScrapeOnlyOwnNode())
		assert.True(t, NodeAgentType.ScrapeOnlyOwnNode())
		assert.False(t, ClusterAgentType.ScrapeOnlyOwnNode())
	})
}
