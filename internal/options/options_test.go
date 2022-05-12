package options

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestDaemonFlag(t *testing.T) {
	t.Run("when all flags are omitted, sets the the options to the defaults", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "all")
	})

	t.Run("when all daemon is true, scrapes the cluster and scrapes its own node", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=true"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "own")
	})

	t.Run("when all daemon is explicitly false, scrapes the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--daemon=false"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "all")
	})

	t.Run("when all scrape-cluster is true, scrapes the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--scrape-cluster=true"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "all")
	})

	t.Run("when all scrape-cluster is false, does not scrape the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--scrape-cluster=false"}))
		assert.Equal(t, opts.ScrapeCluster, false)
		assert.Equal(t, opts.ScrapeNodes, "all")
	})

	t.Run("when all scrape-nodes is all, scrapes the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--scrape-nodes=all"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "all")
	})

	t.Run("when all scrape-nodes is all, scrapes the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--scrape-nodes=own"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "own")
	})

	t.Run("when all scrape-nodes is all, scrapes the cluster and scrapes all nodes", func(t *testing.T) {
		fs := pflag.NewFlagSet("fake-collector", pflag.ContinueOnError)
		opts := NewCollectorRunOptions()

		assert.Nil(t, opts.Parse(fs, []string{"--scrape-nodes=none"}))
		assert.Equal(t, opts.ScrapeCluster, true)
		assert.Equal(t, opts.ScrapeNodes, "none")
	})

	t.Run("when daemon is set with scrape-nodes or scrape-cluster, returns an error", func(t *testing.T) {
		opts := NewCollectorRunOptions()

		expectedErrMsg := "cannot set daemon with either scrape-nodes or scrape-cluster"
		flagCombos := [][]string{
			{"--daemon=true", "--scrape-nodes=all"},
			{"--daemon=true", "--scrape-nodes=own"},
			{"--daemon=true", "--scrape-nodes=none"},
			{"--daemon=true", "--scrape-cluster=true"},
			{"--daemon=true", "--scrape-cluster=false"},
			{"--daemon=false", "--scrape-nodes=none"},
			{"--daemon=false", "--scrape-nodes=own"},
			{"--daemon=false", "--scrape-nodes=all"},
			{"--daemon=false", "--scrape-cluster=true"},
			{"--daemon=false", "--scrape-cluster=false"},
		}
		for _, flagCombo := range flagCombos {
			assert.Errorf(t, opts.Parse(pflag.NewFlagSet("fake-collector", pflag.ContinueOnError), flagCombo), expectedErrMsg)
		}
	})

    t.Run("when scrape-cluster is set to false and scrape-nodes is set to none, returns an error", func(t *testing.T) {
        opts := NewCollectorRunOptions()

        expectedErrMsg := "cannot set scrape-nodes to none with scrape-cluster false"
        flagCombo := []string{"--scrape-cluster=false", "--scrape-nodes=none"}

        assert.Errorf(t, opts.Parse(pflag.NewFlagSet("fake-collector", pflag.ContinueOnError), flagCombo), expectedErrMsg)

    })
}
