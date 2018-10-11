package client

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"time"
)

type WavefrontProxyClient struct {
	proxyAddress string
	conn         net.Conn
}

func NewWavefrontProxyClient(proxyAddress string) WavefrontMetricSender {
	glog.Infof("Creating Wavefront Proxy Client: %s", proxyAddress)
	client := &WavefrontProxyClient{proxyAddress: proxyAddress}
	return client
}

func (c *WavefrontProxyClient) Connect() error {
	var err error
	c.conn, err = net.DialTimeout("tcp", c.proxyAddress, time.Second*10)
	if err != nil {
		glog.Warningf("Unable to connect to Wavefront proxy at address: %s", c.proxyAddress)
		return err
	} else {
		glog.Infof("Connected to Wavefront proxy at address: %s", c.proxyAddress)
		return nil
	}
}

func (c *WavefrontProxyClient) SendMetric(name, value, ts, source, tagStr string) {
	line := fmt.Sprintf("%s %s %s source=\"%s\" %s\n", name, value, ts, source, tagStr)
	c.sendLine(line)
}

func (c *WavefrontProxyClient) sendLine(line string) {
	// if the connection was closed or interrupted - don't cause a panic (we'll retry at next interval)
	defer func() {
		if r := recover(); r != nil {
			// we couldn't write the line so something is wrong with the connection
			c.conn = nil
		}
	}()

	if c.conn != nil {
		c.conn.Write([]byte(line))
	}
}

func (c *WavefrontProxyClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
