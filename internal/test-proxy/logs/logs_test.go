package logs_test

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/test-proxy/logs"
	"testing"
)

func TestLogFormat(t *testing.T) {
	t.Run("logs are in json_array format", func(t *testing.T) {
		jsonArray := `[
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"},
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					]`
		logs.VerifyJsonArray(jsonArray)
	})

	t.Run("logs are in json lines format", func(t *testing.T) {
		jsonLines := `{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}`
		logs.VerifyJsonLines(jsonLines)
	})
}
