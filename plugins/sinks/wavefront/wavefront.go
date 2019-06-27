package wavefront

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
)

const (
	proxyClient  = 1
	directClient = 2
)

var (
	excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}
	sentPoints     gm.Counter
	errPoints      gm.Counter
	msCount        gm.Counter
	filteredPoints gm.Counter
	clientType     gm.Gauge
	sanitizedChars = strings.NewReplacer("+", "-")
)

func init() {
	sentPoints = gm.GetOrRegisterCounter("wavefront.points.sent.count", gm.DefaultRegistry)
	errPoints = gm.GetOrRegisterCounter("wavefront.points.errors.count", gm.DefaultRegistry)
	msCount = gm.GetOrRegisterCounter("wavefront.points.metric-sets.count", gm.DefaultRegistry)
	filteredPoints = gm.GetOrRegisterCounter("wavefront.points.filtered.count", gm.DefaultRegistry)
	clientType = gm.GetOrRegisterGauge("wavefront.sender.type", gm.DefaultRegistry)
}

type wavefrontSink struct {
	WavefrontClient   senders.Sender
	ClusterName       string
	Prefix            string
	globalTags        map[string]string
	filters           filter.Filter
	testMode          bool
	testReceivedLines []string
}

func (sink *wavefrontSink) Name() string {
	return "Wavefront Sink"
}

func (sink *wavefrontSink) Stop() {
	sink.WavefrontClient.Close()
}

func (sink *wavefrontSink) sendPoint(metricName string, value float64, ts int64, source string, tags map[string]string) {
	metricName = sanitizedChars.Replace(metricName)
	if sink.filters != nil && !sink.filters.Match(metricName, tags) {
		filteredPoints.Inc(1)
		glog.V(5).Infof("dropping metric: %s", metricName)
		return
	}

	tags = combineGlobalTags(tags, sink.globalTags)

	if sink.testMode {
		tagStr := ""
		for k, v := range tags {
			tagStr += k + "=\"" + v + "\" "
		}
		line := fmt.Sprintf("%s %f %d source=\"%s\" %s\n", metricName, value, ts, source, tagStr)
		sink.testReceivedLines = append(sink.testReceivedLines, line)
		glog.Infoln(line)
		return
	}
	err := sink.WavefrontClient.SendMetric(metricName, value, ts, source, tags)
	if err != nil {
		errPoints.Inc(1)
		glog.Errorf("error=%q sending metric=%s", err, metricName)
	} else {
		sentPoints.Inc(1)
	}
}

func combineGlobalTags(tags, globalTags map[string]string) map[string]string {
	if tags == nil || len(tags) == 0 {
		return globalTags
	}
	if globalTags == nil || len(globalTags) == 0 {
		return tags
	}

	for k, v := range globalTags {
		// add global tag if key is missing from tags
		if _, exists := tags[k]; !exists {
			tags[k] = v
		}
	}
	return tags
}

func (sink *wavefrontSink) send(batch *metrics.DataBatch) {
	if len(batch.MetricPoints) > 0 {
		sink.processMetricPoints(batch.MetricPoints)
	}
}

func (sink *wavefrontSink) processMetricPoints(points []*metrics.MetricPoint) {
	glog.V(2).Infof("received metric points: %d", len(points))
	for _, point := range points {
		if point.Tags == nil {
			point.Tags = make(map[string]string, 1)
		}
		point.Tags["cluster"] = sink.ClusterName
		sink.sendPoint(point.Metric, point.Value, point.Timestamp, point.Source, point.Tags)
	}
}

func (sink *wavefrontSink) ExportData(batch *metrics.DataBatch) {
	if sink.testMode {
		//clear lines from last batch
		sink.testReceivedLines = sink.testReceivedLines[:0]
		sink.send(batch)
		return
	}
	sink.send(batch)
}

func NewWavefrontSink(uri *url.URL) (metrics.DataSink, error) {
	if len(uri.Scheme) > 0 {
		return nil, fmt.Errorf("scheme should not be set for Wavefront sink")
	}

	if len(uri.Host) > 0 {
		return nil, fmt.Errorf("host should not be set for Wavefront sink")
	}

	storage := &wavefrontSink{
		ClusterName: "k8s-cluster",
		testMode:    false,
	}

	vals := uri.Query()

	if len(vals["proxyAddress"]) > 0 {
		s := strings.Split(vals["proxyAddress"][0], ":")
		host, portStr := s[0], s[1]
		port, err := strconv.Atoi(portStr)

		if err != nil {
			return nil, fmt.Errorf("error parsing proxy port: %s", err.Error())
		}
		storage.WavefrontClient, err = senders.NewProxySender(&senders.ProxyConfiguration{
			Host:        host,
			MetricsPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating proxy sender: %s", err.Error())
		}
		clientType.Update(proxyClient)
	}

	if len(vals["server"]) > 0 {
		server := vals["server"][0]
		if len(vals["token"]) == 0 {
			return nil, fmt.Errorf("token missing for Wavefront sink")
		}
		token := vals["token"][0]

		var err error
		storage.WavefrontClient, err = senders.NewDirectSender(&senders.DirectConfiguration{
			Server: server,
			Token:  token,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating direct sender: %s", err.Error())
		}
		clientType.Update(directClient)
	}

	if storage.WavefrontClient == nil {
		return nil, fmt.Errorf("proxyAddress or server property required for Wavefront sink")
	}

	storage.globalTags = flags.DecodeTags(vals)

	if len(vals["clusterName"]) > 0 {
		storage.ClusterName = vals["clusterName"][0]
	}
	if len(vals["prefix"]) > 0 {
		storage.Prefix = vals["prefix"][0]
	}

	storage.filters = filter.FromQuery(vals)

	if len(vals["testMode"]) > 0 {
		testMode, err := strconv.ParseBool(vals["testMode"][0])
		if err != nil {
			glog.Warning("Unable to parse the testMode argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.testMode = testMode
	}
	return storage, nil
}
