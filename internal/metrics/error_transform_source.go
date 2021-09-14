package metrics

type errorTransformSource struct {
	src     MetricsSource
	errFunc func(err error) error
}

func (c *errorTransformSource) Name() string {
	return c.src.Name()
}

func (c *errorTransformSource) ScrapeMetrics() (*DataBatch, error) {
	dataBatch, err := c.src.ScrapeMetrics()
	return dataBatch, c.errFunc(err)
}

func (c *errorTransformSource) Cleanup() {
	c.src.Cleanup()
}

func NewErrorTransformSource(src MetricsSource, errFunc func(err error) error) MetricsSource {
	return &errorTransformSource{src: src, errFunc: errFunc}
}
