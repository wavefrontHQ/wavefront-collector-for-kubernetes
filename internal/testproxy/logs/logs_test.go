package logs_test

import (
	"encoding/json"
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

func convertMapToJSON(input map[string]interface{}) (string, error) {
	output, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
