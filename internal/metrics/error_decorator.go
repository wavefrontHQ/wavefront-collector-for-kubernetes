package metrics

type errorSourceDecorator struct {
	src     MetricsSource
	errFunc func(err error) error
}

func (c *errorSourceDecorator) Name() string {
	return c.src.Name()
}

func (c *errorSourceDecorator) ScrapeMetrics() (*DataBatch, error) {
	dataBatch, err := c.src.ScrapeMetrics()
	return dataBatch, c.errFunc(err)
}

func (c *errorSourceDecorator) Cleanup() {
	c.src.Cleanup()
}

// NewErrorDecorator creates a MetricSource that transforms ScrapeMetrics errors
func NewErrorDecorator(src MetricsSource, errFunc func(err error) error) MetricsSource {
	return &errorSourceDecorator{src: src, errFunc: errFunc}
}
