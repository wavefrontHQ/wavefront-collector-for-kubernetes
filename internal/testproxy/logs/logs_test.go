package logs_test

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
)

func TestLogFormat(t *testing.T) {
	t.Run("logs are in json_array format", func(t *testing.T) {
		jsonArray := "[" + strings.Join([]string{validLogLine(), validLogLine()}, ",\n") + "]"
		formatValid, _ := logs.VerifyJsonArrayFormat(jsonArray)
		require.True(t, formatValid)
	})

	t.Run("logs are not in json_array format", func(t *testing.T) {
		jsonArray := `{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}`
		formatValid, _ := logs.VerifyJsonArrayFormat(jsonArray)
		require.False(t, formatValid)
	})

	t.Run("logs are in json lines format", func(t *testing.T) {
		jsonLines := strings.Join([]string{validLogLine(), validLogLine()}, "\n")
		formatValid, _ := logs.VerifyJsonLinesFormat(jsonLines)
		require.True(t, formatValid)
	})

	t.Run("logs are not in json lines format", func(t *testing.T) {
		jsonLines := strings.Join([]string{validLogLine(), validLogLine()}, ",\n")
		formatValid, _ := logs.VerifyJsonLinesFormat(jsonLines)
		require.False(t, formatValid)
	})
}

func TestLogTags(t *testing.T) {
	t.Run("logs in json_array format missing tags", func(t *testing.T) {
		logJson := "[" + strings.Join([]string{logLineMissingTags(), logLineMissingTags()}, ",\n") + "]"
		formatValid, logLines := logs.VerifyJsonArrayFormat(logJson)
		require.True(t, formatValid)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.False(t, tagsValid)
		require.ElementsMatch(t, []string{"application", "source", "pod_name", "container_id", "namespace_name", "pod_id", "user_defined_tag", "service", "cluster", "timestamp", "container_name"}, missingTags)
	})

	t.Run("logs in json lines format missing tags", func(t *testing.T) {
		jsonLines := strings.Join([]string{logLineMissingTags(), logLineMissingTags()}, "\n")
		formatValid, logLines := logs.VerifyJsonLinesFormat(jsonLines)
		require.True(t, formatValid)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.False(t, tagsValid)
		require.ElementsMatch(t, []string{"application", "source", "pod_name", "container_id", "namespace_name", "pod_id", "user_defined_tag", "service", "cluster", "timestamp", "container_name"}, missingTags)
	})

	t.Run("logs in json_array format no missing tags", func(t *testing.T) {
		logJson := "[" + strings.Join([]string{validLogLine(), validLogLine()}, ",\n") + "]"
		formatValid, logLines := logs.VerifyJsonArrayFormat(logJson)
		require.True(t, formatValid)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.True(t, tagsValid)
		require.Nil(t, missingTags)
	})

	t.Run("logs in json lines format no missing tags", func(t *testing.T) {
		logJson := strings.Join([]string{validLogLine(), validLogLine()}, "\n")
		formatValid, logLines := logs.VerifyJsonLinesFormat(logJson)
		require.True(t, formatValid)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.True(t, tagsValid)
		require.Nil(t, missingTags)
	})
}

func validLogLine() string {
	return "{\"user_defined_tag\":\"some-value\",\"cluster\":\"testk8scluster\",\"message\":\"testlogmessage\",\"service\":\"none\", \"application\":\"none\", \"source\":\"none\"," +
		"\"timestamp\":\"none\",\"pod_name\":\"none\",\"container_name\":\"none\",\"namespace_name\":\"none\",\"pod_id\":\"none\",\"container_id\":\"none\",\"timestamp\":1.672782310258e+12}"
}

func logLineMissingTags() string {
	return "{\"message\":\"testlogmessage\"}"
}
