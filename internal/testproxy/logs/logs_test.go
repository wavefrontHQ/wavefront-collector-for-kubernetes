package logs_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
)

var (
	validLogMap       = map[string]interface{}{"application": "none", "cluster": "testk8scluster", "container_id": "none", "container_name": "none", "message": "testlogmessage", "namespace_name": "none", "pod_id": "none", "pod_name": "none", "service": "none", "source": "none", "timestamp": 1.672782310258e+12, "user_defined_tag": "some-value"}
	missingTagsLogMap = map[string]interface{}{"message": "testlogmessage"}
)

func TestLogFormat(t *testing.T) {
	t.Run("logs are in json_array format", func(t *testing.T) {
		logMap, err := convertMapToJSON(validLogMap)
		require.Nil(t, err)

		jsonArray := "[" + strings.Join([]string{logMap, logMap}, ",\n") + "]"
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
		logMap, err := convertMapToJSON(validLogMap)
		require.Nil(t, err)

		jsonLines := strings.Join([]string{logMap, logMap}, "\n")
		formatValid, _ := logs.VerifyJsonLinesFormat(jsonLines)
		require.True(t, formatValid)
	})

	t.Run("logs are not in json lines format", func(t *testing.T) {
		logMap, err := convertMapToJSON(validLogMap)
		require.Nil(t, err)

		jsonLines := strings.Join([]string{logMap, logMap}, ",\n")
		formatValid, _ := logs.VerifyJsonLinesFormat(jsonLines)
		require.False(t, formatValid)
	})
}

func TestValidateTags(t *testing.T) {
	t.Run("json_array format missing tags", func(t *testing.T) {
		var logLines []interface{}
		logLines = append(logLines, missingTagsLogMap)
		logLines = append(logLines, missingTagsLogMap)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.False(t, tagsValid)
		require.ElementsMatch(t, []string{"application", "source", "pod_name", "container_id", "namespace_name", "pod_id", "user_defined_tag", "service", "cluster", "timestamp", "container_name"}, missingTags)
	})

	t.Run("json lines format missing tags", func(t *testing.T) {
		var logLines []interface{}
		logLines = append(logLines, missingTagsLogMap)
		logLines = append(logLines, missingTagsLogMap)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.False(t, tagsValid)
		require.ElementsMatch(t, []string{"application", "source", "pod_name", "container_id", "namespace_name", "pod_id", "user_defined_tag", "service", "cluster", "timestamp", "container_name"}, missingTags)
	})

	t.Run("json_array format no missing tags", func(t *testing.T) {
		var logLines []interface{}
		logLines = append(logLines, validLogMap)
		logLines = append(logLines, validLogMap)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.True(t, tagsValid)
		require.Nil(t, missingTags)
	})

	t.Run("json lines format no missing tags", func(t *testing.T) {
		var logLines []interface{}
		logLines = append(logLines, validLogMap)
		logLines = append(logLines, validLogMap)

		tagsValid, missingTags := logs.ValidateTags(logLines)
		require.True(t, tagsValid)
		require.Nil(t, missingTags)
	})
}

func convertMapToJSON(input map[string]interface{}) (string, error) {
	output, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
