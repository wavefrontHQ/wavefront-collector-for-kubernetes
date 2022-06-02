package options

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestDaemonFlag(t *testing.T) {
	t.Run("when all flags are omitted, defaults to agentType=all (deployement)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{}))
		assert.Equal(t, "all",  opts.agentType.String())
	})

	t.Run("when only daemon is true, deploy agentType=legacy (daemonset with leader)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=true"}))
		assert.Equal(t, opts.agentType.String(), "legacy")
	})

	t.Run("when daemon is explicitly false, deploy agentType=all (deployment)", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=false"}))
        assert.Equal(t, opts.agentType.String(), "all")
	})

    t.Run("--agent=all is a valid option", func(t *testing.T) {
        fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
        opts := NewCollectorRunOptions()

        assert.Nil(t, opts.Parse(fs, []string{"--agent=all"}))
        assert.Equal(t, opts.agentType.String(), "all")
    })

    t.Run("--agent=legacy is a valid option", func(t *testing.T) {
        fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
        opts := NewCollectorRunOptions()

        assert.Nil(t, opts.Parse(fs, []string{"--agent=legacy"}))
        assert.Equal(t, opts.agentType.String(), "legacy")
    })

	t.Run("--agent=node is a valid option", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--agent=node"}))
		assert.Equal(t, opts.agentType.String(), "node")
	})

    t.Run("--agent=cluster is a valid option", func(t *testing.T) {
        fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
        opts := NewCollectorRunOptions()

        assert.Nil(t, opts.Parse(fs, []string{"--agent=cluster"}))
        assert.Equal(t, opts.agentType.String(), "cluster")
    })

	t.Run("When both --daemon and --agent are set, returns an error", func(t *testing.T) {
		opts := NewCollectorRunOptions()

		expectedErrMsg := "cannot set --daemon with --agent"
		flagCombos := [][]string{
			{"--daemon=true", "--agent=legacy"},
			{"--daemon=false", "--agent=node"},
		}
		for _, flagCombo := range flagCombos {
			assert.Errorf(t, opts.Parse(pflag.NewFlagSet("fake-collector", pflag.ContinueOnError), flagCombo), expectedErrMsg)
		}
	})
}
