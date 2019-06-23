package telegraf

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"
)

func initPlugin(input telegraf.Input, vals map[string][]string) error {
	if _, ok := vals["tg.conf"]; !ok {
		return fmt.Errorf("missing telegraf configuration")
	}

	conf := vals["tg.conf"][0]
	if err := toml.Unmarshal([]byte(conf), input); err != nil {
		return err
	}
	return nil
}
