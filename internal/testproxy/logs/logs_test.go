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
		logMap := map[string]interface{}{"some-expected-tag": "some-value"}
		expectedTag := []string{"some-expected-tag"}

		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.True(t, tagsValid)
		require.Nil(t, missingTags)
		require.Nil(t, emptyTags)
	})

	t.Run("expected tags are not found", func(t *testing.T) {
		logMap := map[string]interface{}{"wrong-expected-tag": "some-value"}
		expectedTag := []string{"some-expected-tag"}

		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.False(t, tagsValid)
		require.Contains(t, missingTags, "some-expected-tag")
		require.Nil(t, emptyTags)
	})

	t.Run("expected tags are found but the value is nil", func(t *testing.T) {
		logMap := map[string]interface{}{"some-expected-tag": nil}
		expectedTag := []string{"some-expected-tag"}

		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.False(t, tagsValid)
		require.Nil(t, missingTags)
		require.Contains(t, emptyTags, "some-expected-tag")
	})

	t.Run("expected tags are found but the value is empty", func(t *testing.T) {
		logMap := map[string]interface{}{"some-expected-tag": ""}
		expectedTag := []string{"some-expected-tag"}

		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.False(t, tagsValid)
		require.Nil(t, missingTags)
		require.Contains(t, emptyTags, "some-expected-tag")
	})

	t.Run("expected tags are found but the value is not nil or an empty string", func(t *testing.T) {
		logMap := map[string]interface{}{"some-expected-tag": 0}
		expectedTag := []string{"some-expected-tag"}

		var logLines []interface{}
		logLines = append(logLines, logMap)

		logVerifier := logs.NewLogVerifier(expectedTag, nil, nil)
		tagsValid, missingTags, emptyTags := logVerifier.ValidateExpectedTags(logLines)

		require.True(t, tagsValid)
		require.Nil(t, missingTags)
		require.Nil(t, emptyTags)
	})
}

func TestValidateAllowedTags(t *testing.T) {
	t.Run("valid if there is a tag from the allowed list", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-allow": "some-value"}
		tagAllowList := map[string][]string{"tag-to-allow": {"some-value"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedTags := logVerifier.ValidateAllowedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})

	t.Run("invalid if there are no tags from the allowed list", func(t *testing.T) {
		tagMap := map[string]interface{}{"some-key": "some-value"}
		tagAllowList := map[string][]string{"tag-to-allow": {"value-to-allow"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedLogs := logVerifier.ValidateAllowedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, logLines, unexpectedLogs)
	})

	t.Run("invalid if the tag key is allowed but the tag value is not allowed", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-allow": "some-value"}
		tagAllowList := map[string][]string{"tag-to-allow": {"value-to-allow"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedLogs := logVerifier.ValidateAllowedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, logLines, unexpectedLogs)
	})

	t.Run("invalid if the tag value is allowed but the tag key is not allowed", func(t *testing.T) {
		tagMap := map[string]interface{}{"some-key": "value-to-allow"}
		tagAllowList := map[string][]string{"tag-to-allow": {"value-to-allow"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedLogs := logVerifier.ValidateAllowedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, logLines, unexpectedLogs)
	})

	t.Run("valid when all logs have a tag from the allow list", func(t *testing.T) {
		validLog1 := map[string]interface{}{"tag-to-allow": "value-to-allow"}
		validLog2 := map[string]interface{}{"other-tag-to-allow": "other-value-to-allow"}
		tagAllowList := map[string][]string{
			"tag-to-allow":       {"value-to-allow"},
			"other-tag-to-allow": {"other-value-to-allow"},
		}

		var logLines []interface{}
		logLines = append(logLines, validLog1)
		logLines = append(logLines, validLog2)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedLogs := logVerifier.ValidateAllowedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedLogs)
	})

	t.Run("invalid if there is one log in the list that does not have any allowed tags", func(t *testing.T) {
		invalidLog := map[string]interface{}{"some-key": "some-tag"}
		validLog := map[string]interface{}{"tag-to-allow": "value-to-allow"}
		tagAllowList := map[string][]string{"tag-to-allow": {"value-to-allow"}}

		var logLines []interface{}
		logLines = append(logLines, invalidLog)
		logLines = append(logLines, validLog)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedLogs := logVerifier.ValidateAllowedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, invalidLog, unexpectedLogs[0])
	})

	t.Run("validate allow list filter tags when the allowed list tag value is an empty list", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-allow": "some-value"}
		tagAllowList := map[string][]string{"tag-to-allow": {}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, tagAllowList, nil)
		tagsAreValid, unexpectedTags := logVerifier.ValidateAllowedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})
}

func TestValidateDeniedTags(t *testing.T) {
	t.Run("valid if no denied tags are present", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-allow": "some-value"}
		tagDenyList := map[string][]string{"tag-to-deny": {"some-value"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})

	t.Run("invalid if denied tags are present and returns which ones were unexpected", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-deny": "some-value", "tag-to-allow": "some-value"}
		tagDenyList := map[string][]string{"tag-to-deny": {"some-value"}}
		unexpectedTag := map[string]interface{}{"tag-to-deny": "some-value"}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, unexpectedTags, unexpectedTag)
	})

	t.Run("valid if the tag key is on the deny list but the tag value is not", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-key-to-deny": "tag-value-to-allow"}
		tagDenyList := map[string][]string{"tag-key-to-deny": {"tag-value-to-deny"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})

	t.Run("valid if the tag value is on the deny list but the tag key is not", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-key-to-allow": "tag-value-to-deny"}
		tagDenyList := map[string][]string{"tag-key-to-deny": {"tag-value-to-deny"}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})

	t.Run("valid when all logs have tags that are not in the deny list", func(t *testing.T) {
		validLog1 := map[string]interface{}{"tag-to-allow": "value-to-allow"}
		validLog2 := map[string]interface{}{"other-tag-to-allow": "other-value-to-allow"}
		tagDenyList := map[string][]string{
			"tag-to-deny":       {"value-to-deny"},
			"other-tag-to-deny": {"other-value-to-deny"},
		}

		var logLines []interface{}
		logLines = append(logLines, validLog1)
		logLines = append(logLines, validLog2)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.True(t, tagsAreValid)
		require.Empty(t, unexpectedTags)
	})

	t.Run("invalid if at least one log has a tag in the deny list", func(t *testing.T) {
		invalidTagMap := map[string]interface{}{"tag-key-to-deny": "tag-value-to-deny"}
		validTagMap := map[string]interface{}{"tag-key-to-allow": "tag-value-to-allow"}
		tagDenyList := map[string][]string{"tag-key-to-deny": {"tag-value-to-deny"}}

		var logLines []interface{}
		logLines = append(logLines, invalidTagMap)
		logLines = append(logLines, validTagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, invalidTagMap, unexpectedTags)
	})

	t.Run("validate deny list filter tags when the deny list tag value is an empty list", func(t *testing.T) {
		tagMap := map[string]interface{}{"tag-to-deny": "some-value"}
		tagDenyList := map[string][]string{"tag-to-deny": {}}

		var logLines []interface{}
		logLines = append(logLines, tagMap)

		logVerifier := logs.NewLogVerifier(nil, nil, tagDenyList)
		tagsAreValid, unexpectedTags := logVerifier.ValidateDeniedTags(logLines)

		require.False(t, tagsAreValid)
		require.Equal(t, tagMap, unexpectedTags)
	})
}
