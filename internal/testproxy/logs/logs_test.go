package logs_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
)

func TestVerifyJsonArrayFormat(t *testing.T) {
	t.Run("valid when json array is in expected format", func(t *testing.T) {
		jsonArray := `[{"key1":"value1","key2":"value2"},{"key3":"value3","key4":"value4"}]`
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonArrayFormat([]byte(jsonArray))

		require.True(t, formatValid)
	})

	t.Run("invalid when json array is empty brackets", func(t *testing.T) {
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonArrayFormat([]byte("[]"))

		require.False(t, formatValid)
	})

	t.Run("invalid when json array is json lines format", func(t *testing.T) {
		jsonLines := `{"key1":"value1", "key2":"value2"}
					{"key3":"value3", "key4":"value4"}`
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonArrayFormat([]byte(jsonLines))

		require.False(t, formatValid)
	})

	t.Run("invalid when json array is empty", func(t *testing.T) {
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonArrayFormat([]byte{})

		require.False(t, formatValid)
	})

	t.Run("invalid when json array is not a json array", func(t *testing.T) {
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonArrayFormat([]byte("{}"))

		require.False(t, formatValid)
	})
}

func TestVerifyJsonLinesFormat(t *testing.T) {
	t.Run("valid when json lines is in expected format", func(t *testing.T) {
		jsonLines := `{"key1":"value1", "key2":"value2"}
					{"key3":"value3", "key4":"value4"}`
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonLinesFormat([]byte(jsonLines))

		require.True(t, formatValid)
	})

	t.Run("invalid when json lines is in invalid json lines format with comma between elements", func(t *testing.T) {
		jsonArray := `{"key1":"value1","key2":"value2"},
						{"key3":"value3","key4":"value4"}`
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonLinesFormat([]byte(jsonArray))

		require.False(t, formatValid)
	})

	t.Run("invalid when json lines is in json array format", func(t *testing.T) {
		jsonArray := `[{"key1":"value1","key2":"value2"},
						{"key3":"value3","key4":"value4"}]`
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonLinesFormat([]byte(jsonArray))

		require.False(t, formatValid)
	})

	t.Run("invalid when json lines data is empty", func(t *testing.T) {
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonLinesFormat([]byte{})

		require.False(t, formatValid)
	})

	t.Run("invalid when json lines are empty", func(t *testing.T) {
		logVerifier := logs.NewLogVerifier(nil, nil, nil)

		formatValid, _ := logVerifier.VerifyJsonLinesFormat([]byte("{}"))

		require.False(t, formatValid)
	})
}

func TestValidateExpectedTags(t *testing.T) {
	t.Run("all expected tags are found and are not empty", func(t *testing.T) {
		expectedTag := []string{"some-expected-tag"}
		logMap := map[string]interface{}{
			"some-expected-tag": "some-value",
		}
		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.True(t, tagsValid)
		require.Nil(t, missingTags)
		require.Nil(t, emptyTags)
	})
	// TODO: if expected tags are not found
	// TODO: if expected tags are found and the value is empty
}
