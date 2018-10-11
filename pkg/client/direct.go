package client

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
	"sync"
	"time"
)

type WavefrontDirectClient struct {
	reporter      Reporter
	points        []string
	mtx           sync.Mutex
	flushTicker   *time.Ticker
	maxFlushSize  int
	maxBufferSize int
}

func NewWavefrontDirectClient(server, token string) WavefrontMetricSender {
	glog.Infof("Creating Wavefront DirectIngestion Client : %s", server)
	reporter := NewDirectReporter(server, token)
	client := &WavefrontDirectClient{
		reporter:      reporter,
		flushTicker:   time.NewTicker(time.Second * time.Duration(1)),
		maxFlushSize:  40000,
		maxBufferSize: 100000,
	}
	go client.flush()
	return client
}

func (c *WavefrontDirectClient) SendMetric(name, value, ts, source, tagStr string) {
	line := fmt.Sprintf("%s %s %s source=\"%s\" %s\n", name, value, ts, source, tagStr)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.points = append(c.points, line)
}

func (c *WavefrontDirectClient) Connect() error {
	// no-op
	return nil
}

func (c *WavefrontDirectClient) Close() {
	// no-op
}

func (c *WavefrontDirectClient) flush() {
	for range c.flushTicker.C {
		c.flushPoints(c.getPointsBatch())
	}
}

func (c *WavefrontDirectClient) flushPoints(points []string) {
	if len(points) == 0 {
		return
	}
	pointLines := strings.Join(points, "\n")
	resp, err := c.reporter.Report("wavefront", pointLines)
	if err != nil {
		glog.Errorf("Error reporting points to Wavefront: %q", err)
		return
	}
	if resp.StatusCode >= 300 {
		glog.Errorf("Error reporting points to Wavefront: %d", resp.StatusCode)
	}
	glog.V(2).Infof("successfully reported %d points to Wavefront: %d", len(points), resp.StatusCode)
}

func (c *WavefrontDirectClient) getPointsBatch() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	currLen := len(c.points)
	batchSize := min(currLen, c.maxFlushSize)
	batchPoints := c.points[:batchSize]
	c.points = c.points[batchSize:currLen]
	return batchPoints
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
