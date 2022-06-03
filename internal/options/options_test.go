package options

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestDaemonFlag(t *testing.T) {
	t.Run("when all flags are omitted, defaults to AgentType=all (deployement)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{}))
		assert.Equal(t, AllAgentType, opts.AgentType)
	})

	t.Run("when only daemon is true, sets AgentType=legacy (daemonset with leader)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=true"}))
		assert.Equal(t, LegacyAgentType, opts.AgentType)
	})

	t.Run("when daemon is explicitly false, deploy AgentType=all (deployment)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=false"}))
		assert.Equal(t, AllAgentType, opts.AgentType)
	})

	t.Run("--agent=all is a valid option", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--agent=all"}))
		assert.Equal(t, AllAgentType, opts.AgentType)
	})

	t.Run("--agent=legacy is a valid option", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--agent=legacy"}))
		assert.Equal(t, LegacyAgentType, opts.AgentType)
	})

	t.Run("--agent=node is a valid option", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--agent=node"}))
		assert.Equal(t, NodeAgentType, opts.AgentType)
	})

	t.Run("--agent=cluster is a valid option", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--agent=cluster"}))
		assert.Equal(t, ClusterAgentType, opts.AgentType)
	})

	t.Run("returns an error when both --daemon and --agent are set", func(t *testing.T) {
		opts := NewCollectorRunOptions()

		flagCombos := [][]string{
			{"--daemon=true", "--agent=legacy"},
			{"--daemon=false", "--agent=node"},
		}
		for _, flagCombo := range flagCombos {
			assert.Errorf(t, opts.Parse(pflag.NewFlagSet("fake-collector", pflag.ContinueOnError), flagCombo), DaemonAndAgentErr.Error())
		}
	})

	t.Run("validates --agent as an enum", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Errorf(t, opts.Parse(fs, []string{"--agent=invalid"}), InvalidAgentTypeErr.Error())
	})
}
