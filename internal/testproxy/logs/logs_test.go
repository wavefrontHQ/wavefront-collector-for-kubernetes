package logs_test

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
)

func TestLogFormat(t *testing.T) {
	t.Run("logs are in json_array format with expected tags", func(t *testing.T) {
		require.True(t, logs.VerifyJsonArray(validJsonArray()))
	})

	t.Run("logs are not in json_array format", func(t *testing.T) {
		jsonArray := `{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}`
		require.False(t, logs.VerifyJsonArray(jsonArray))
	})

	t.Run("logs are in json lines format with expected tags", func(t *testing.T) {
		jsonLines := strings.Join([]string{validLogLine(), validLogLine()}, "\n")
		require.True(t, logs.VerifyJsonLines(jsonLines))
	})

	t.Run("logs are not in json lines format", func(t *testing.T) {
		jsonLines := strings.Join([]string{validLogLine(), validLogLine()}, ",\n")
		require.False(t, logs.VerifyJsonLines(jsonLines))
	})
}

func validJsonArray() string {
	return "[" + strings.Join([]string{validLogLine(), validLogLine()}, ",\n") + "]"
}

func validLogLine() string {
	return "{\"cluster\":\"testk8scluster\",\"message\":\"testlogmessage\",\"service\":\"none\", \"application\":\"none\", \"source\":\"none\"," +
		"\"timestamp\":\"none\",\"pod_name\":\"none\",\"container_name\":\"none\",\"namespace_name\":\"none\",\"pod_id\":\"none\",\"container_id\":\"none\"}"
}
