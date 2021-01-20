package wavefront

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	"sort"
	"sync"
)

type TestSender struct {
	testReceivedLines string
	mutex             sync.Mutex
}

func NewTestSender() senders.Sender {
	log.SetFormatter(&log.JSONFormatter{})
	return &TestSender{
		testReceivedLines: "",
	}
}

func (t *TestSender) SendMetric(name string, value float64, _ int64, _ string, tags map[string]string) error {
	line := fmt.Sprintf("Metric: %s %f %s\n", name, value, orderedTagString(tags))

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.testReceivedLines += line
	log.Infoln(line)

	return nil
}

func (t *TestSender) GetReceivedLines() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.testReceivedLines
}

func (t *TestSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	annotations := map[string]interface{}{}
	annotations["annotations"] = map[string]string{}
	for _, setter := range setters {
		setter(annotations)
	}
	line := fmt.Sprintf("%s %s source=\"%s\" %s\n", name, "event", source, orderedTagString(tags))
	log.Infoln(line)

	return nil
}

func orderedTagString(tags map[string]string) string {
	tagNames := sortKeys(tags)

	tagStr := ""
	for _, tagName := range tagNames {
		if tagName != "cluster" {
			tagStr += tagName + "=\"" + tags[tagName] + "\" "
		}
	}
	return tagStr
}

func sortKeys(tags map[string]string) []string {
	tagCount := len(tags)
	tagNames := make([]string, tagCount)

	count := 0
	for tagName := range tags {
		tagNames[count] = tagName
		count++
	}

	sort.Strings(tagNames)
	return tagNames
}

func (t *TestSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	return nil
}

func (t *TestSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	return nil
}

func (t *TestSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []senders.SpanTag, spanLogs []senders.SpanLog) error {
	return nil
}

func (t *TestSender) Flush() error {
	return nil
}

func (t *TestSender) GetFailureCount() int64 {
	return 0
}

func (t *TestSender) Start() {
}

func (t *TestSender) Close() {
}
