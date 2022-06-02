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
		assert.False(t, ClusterAgentType.ScrapeCluster())
	})

	t.Run("ScrapeNodes", func(t *testing.T) {
		assert.Equal(t, "all", AllAgentType.ScrapeNodes())
		assert.Equal(t, "own", LegacyAgentType.ScrapeNodes())
		assert.Equal(t, "own", NodeAgentType.ScrapeNodes())
		assert.Equal(t, "none", ClusterAgentType.ScrapeNodes())
	})
}
