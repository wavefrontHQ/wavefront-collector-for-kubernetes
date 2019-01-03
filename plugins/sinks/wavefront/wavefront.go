package wavefront

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/golang/glog"
	gm "github.com/rcrowley/go-metrics"
)

const (
	sysSubContainerName = "system.slice/"
	proxyClient         = 1
	directClient        = 2
)

var (
	excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}
	sentPoints     gm.Counter
	errPoints      gm.Counter
	msCount        gm.Counter
	clientType     gm.Gauge
)

func init() {
	sentPoints = gm.GetOrRegisterCounter("wavefront.points.sent.count", gm.DefaultRegistry)
	errPoints = gm.GetOrRegisterCounter("wavefront.points.errors.count", gm.DefaultRegistry)
	msCount = gm.GetOrRegisterCounter("wavefront.points.metric-sets.count", gm.DefaultRegistry)
	clientType = gm.GetOrRegisterGauge("wavefront.sender.type", gm.DefaultRegistry)
}

type wavefrontSink struct {
	WavefrontClient   senders.Sender
	ClusterName       string
	Prefix            string
	IncludeLabels     bool
	IncludeContainers bool
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

func (sink *wavefrontSink) cleanMetricName(metricType string, metricName string) string {
	return sink.Prefix + metricType + "." + strings.Replace(metricName, "/", ".", -1)
}

func (sink *wavefrontSink) addLabelTags(ms *metrics.MetricSet, tags map[string]string) {
	for _, labelName := range sortedLabelKeys(ms.Labels) {
		labelValue := ms.Labels[labelName]
		if labelName == "labels" {
			//only parse labels if IncludeLabels == true
			if sink.IncludeLabels {
				for _, label := range strings.Split(labelValue, ",") {
					//labels = app:webproxy,version:latest
					tagParts := strings.SplitN(label, ":", 2)
					if len(tagParts) == 2 {
						tags["label."+tagParts[0]] = tagParts[1]
					}
				}
			}
		} else {
			tags[labelName] = labelValue
		}
	}
}

func (sink *wavefrontSink) send(batch *metrics.DataBatch) {
	if len(batch.MetricPoints) > 0 {
		sink.processMetricPoints(batch.MetricPoints)
	}
	if len(batch.MetricSets) > 0 {
		sink.processMetricSets(batch.MetricSets, batch.Timestamp)
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

func (sink *wavefrontSink) processMetricSets(metricSets map[string]*metrics.MetricSet, ts time.Time) {
	glog.V(2).Infof("received metric sets: %d", len(metricSets))

	metricCounter := 0

	for _, key := range sortedMetricSetKeys(metricSets) {
		ms := metricSets[key]

		// Populate tag map
		tags := make(map[string]string)
		// Make sure all metrics are tagged with the cluster name
		tags["cluster"] = sink.ClusterName
		// Add pod labels as tags
		sink.addLabelTags(ms, tags)
		metricType := tags["type"]
		if strings.Contains(tags["container_name"], sysSubContainerName) {
			//don't send system subcontainers
			continue
		}
		if sink.IncludeContainers == false && strings.Contains(metricType, "pod_container") {
			// the user doesn't want to include container metrics (only pod and above)
			continue
		}
		for _, metricName := range sortedMetricValueKeys(ms.MetricValues) {
			metricValue := ms.MetricValues[metricName]
			var value float64
			if metrics.ValueInt64 == metricValue.ValueType {
				value = float64(metricValue.IntValue)
			} else if metrics.ValueFloat == metricValue.ValueType {
				value = metricValue.FloatValue
			} else {
				continue
			}

			ts := ts.Unix()
			source := ""
			if metricType == "cluster" {
				source = sink.ClusterName
			} else if metricType == "ns" {
				source = tags["namespace_name"] + "-ns"
			} else {
				source = tags["hostname"]
			}
			processTags(tags)
			sink.sendPoint(sink.cleanMetricName(metricType, metricName), value, ts, source, tags)
			metricCounter = metricCounter + 1
		}
		for _, metric := range ms.LabeledMetrics {
			metricName := sink.cleanMetricName(metricType, metric.Name)
			var value float64
			if metrics.ValueInt64 == metric.ValueType {
				value = float64(metric.IntValue)
			} else if metrics.ValueFloat == metric.ValueType {
				value = metric.FloatValue
			} else {
				continue
			}

			ts := ts.Unix()
			source := tags["hostname"]
			for labelName, labelValue := range metric.Labels {
				tags[labelName] = labelValue
			}
			processTags(tags)
			metricCounter = metricCounter + 1
			sink.sendPoint(metricName, value, ts, source, tags)
		}
	}
	msCount.Inc(int64(metricCounter))
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
		ClusterName:       "k8s-cluster",
		Prefix:            "kubernetes.",
		IncludeLabels:     false,
		IncludeContainers: true,
		testMode:          false,
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

	if len(vals["clusterName"]) > 0 {
		storage.ClusterName = vals["clusterName"][0]
	}
	if len(vals["prefix"]) > 0 {
		storage.Prefix = vals["prefix"][0]
	}
	if len(vals["includeLabels"]) > 0 {
		incLabels := false
		incLabels, err := strconv.ParseBool(vals["includeLabels"][0])
		if err != nil {
			glog.Warning("Unable to parse the includeLabels argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.IncludeLabels = incLabels
	}
	if len(vals["includeContainers"]) > 0 {
		incContainers := false
		incContainers, err := strconv.ParseBool(vals["includeContainers"][0])
		if err != nil {
			glog.Warning("Unable to parse the includeContainers argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.IncludeContainers = incContainers
	}
	if len(vals["testMode"]) > 0 {
		testMode := false
		testMode, err := strconv.ParseBool(vals["testMode"][0])
		if err != nil {
			glog.Warning("Unable to parse the testMode argument. This argument is a boolean, please pass \"true\" or \"false\"")
			return nil, err
		}
		storage.testMode = testMode
	}
	return storage, nil
}
