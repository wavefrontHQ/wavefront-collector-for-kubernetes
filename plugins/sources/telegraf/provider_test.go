package telegraf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoDiscoveredTelegrafPluginSource(t *testing.T) {
	t.Run("static source", func(t *testing.T) {
		ms := newTelegrafPluginSource("", nil, "", map[string]string{}, nil, "")

		assert.False(t, ms.AutoDiscovered(), "telegraf plugin auto-discovery")
	})

	t.Run("discovered source", func(t *testing.T) {
		ms := newTelegrafPluginSource("", nil, "", map[string]string{}, nil, "some-auto-discovery-method")

		assert.True(t, ms.AutoDiscovered(), "telegraf plugin auto-discovery")
	})
}
