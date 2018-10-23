package wavefront

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/pkg/client"
)

const (
	sysSubContainerName = "system.slice/"
)

var excludeTagList = [...]string{"namespace_id", "host_id", "pod_id", "hostname"}

type wavefrontSink struct {
	WavefrontClient   client.WavefrontMetricSender
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

func (sink *wavefrontSink) sendPoint(metricName string, metricValStr string, ts string, source string, tagStr string) {
	if sink.testMode {
		line := fmt.Sprintf("%s %s %s source=\"%s\" %s\n", metricName, metricValStr, ts, source, tagStr)
		sink.testReceivedLines = append(sink.testReceivedLines, line)
		glog.Infoln(line)
		return
	}
	sink.WavefrontClient.SendMetric(metricName, metricValStr, ts, source, tagStr)
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
		metricValStr := fmt.Sprintf("%f", point.Value)
		ts := strconv.FormatInt(point.Timestamp, 10)
		point.Tags["cluster"] = sink.ClusterName
		tagStr := tagsToString(point.Tags)
		sink.sendPoint(point.Metric, metricValStr, ts, point.Source, tagStr)
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
			var metricValStr string
			metricValue := ms.MetricValues[metricName]
			if metrics.ValueInt64 == metricValue.ValueType {
				metricValStr = fmt.Sprintf("%d", metricValue.IntValue)
			} else if metrics.ValueFloat == metricValue.ValueType { // W
				metricValStr = fmt.Sprintf("%f", metricValue.FloatValue)
			} else {
				//do nothing for now
				metricValStr = ""
			}
			if metricValStr != "" {
				ts := strconv.FormatInt(ts.Unix(), 10)
				source := ""
				if metricType == "cluster" {
					source = sink.ClusterName
				} else if metricType == "ns" {
					source = tags["namespace_name"] + "-ns"
				} else {
					source = tags["hostname"]
				}
				tagStr := tagsToString(tags)
				sink.sendPoint(sink.cleanMetricName(metricType, metricName), metricValStr, ts, source, tagStr)
				metricCounter = metricCounter + 1
			}
		}
		for _, metric := range ms.LabeledMetrics {
			metricName := sink.cleanMetricName(metricType, metric.Name)
			metricValStr := ""
			if metrics.ValueInt64 == metric.ValueType {
				metricValStr = fmt.Sprintf("%d", metric.IntValue)
			} else if metrics.ValueFloat == metric.ValueType { // W
				metricValStr = fmt.Sprintf("%f", metric.FloatValue)
			} else {
				//do nothing for now
				metricValStr = ""
			}
			if metricValStr != "" {
				ts := strconv.FormatInt(ts.Unix(), 10)
				source := tags["hostname"]
				tagStr := tagsToString(tags)
				for labelName, labelValue := range metric.Labels {
					tagStr += labelName + "=\"" + labelValue + "\" "
				}
				metricCounter = metricCounter + 1
				sink.sendPoint(metricName, metricValStr, ts, source, tagStr)
			}
		}
	}
}

func (sink *wavefrontSink) ExportData(batch *metrics.DataBatch) {

	if sink.testMode {
		//clear lines from last batch
		sink.testReceivedLines = sink.testReceivedLines[:0]
		sink.send(batch)
		return
	}

	//make sure we're Connected before sending a real batch
	err := sink.connect()
	if err != nil {
		glog.Warning(err)
	}
	sink.send(batch)
}

func (sink *wavefrontSink) connect() error {
	err := sink.WavefrontClient.Connect()
	if err != nil {
		glog.Warning(err.Error())
		return err
	}
	return nil
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
		storage.WavefrontClient = client.NewWavefrontProxyClient(vals["proxyAddress"][0])
	}

	if len(vals["server"]) > 0 {
		server := vals["server"][0]
		if len(vals["token"]) == 0 {
			return nil, fmt.Errorf("token missing for Wavefront sink")
		}
		token := vals["token"][0]
		storage.WavefrontClient = client.NewWavefrontDirectClient(server, token)
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

func tagsToString(tags map[string]string) string {
	tagStr := ""
	for k, v := range tags {
		// ignore tags with empty values as well so the data point doesn't fail validation
		if excludeTag(k) == false && len(v) > 0 {
			tagStr += k + "=\"" + v + "\" "
		}
	}
	return tagStr
}

func excludeTag(a string) bool {
	for _, b := range excludeTagList {
		if b == a {
			return true
		}
	}
	return false
}

func sortedMetricSetKeys(m map[string]*metrics.MetricSet) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedLabelKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedMetricValueKeys(m map[string]metrics.MetricValue) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
