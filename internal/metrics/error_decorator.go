package metrics

type errorSourceDecorator struct {
	src     Source
	errFunc func(err error) error
}

func (c *errorSourceDecorator) Name() string {
	return c.src.Name()
}

func (c *errorSourceDecorator) AutoDiscovered() bool {
	return c.src.AutoDiscovered()
}

func (c *errorSourceDecorator) Scrape() (*Batch, error) {
	dataBatch, err := c.src.Scrape()
	return dataBatch, c.errFunc(err)
}

func (c *errorSourceDecorator) Cleanup() {
	c.src.Cleanup()
}

// NewErrorDecorator creates a MetricSource that transforms Scrape errors
func NewErrorDecorator(src Source, errFunc func(err error) error) Source {
	return &errorSourceDecorator{src: src, errFunc: errFunc}
}
