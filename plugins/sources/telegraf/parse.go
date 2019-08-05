package telegraf

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"
)

func initPlugin(input telegraf.Input, conf string) error {
	if len(conf) == 0 {
		return fmt.Errorf("missing telegraf configuration")
	}
	if err := toml.Unmarshal([]byte(conf), input); err != nil {
		return err
	}
	return nil
}
