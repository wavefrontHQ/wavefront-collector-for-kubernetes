package wavefront

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanTags(t *testing.T) {
	t.Run("excludes tags in the exclude tag list", func(t *testing.T) {
		for _, excludedTagName := range excludeTagList {
			actual := map[string]string{excludedTagName: "some-value"}
			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, map[string]string{}, actual)
		}
	})

	t.Run("excludes tags with given prefixes", func(t *testing.T) {
		for _, excludedTagName := range excludeTagPrefixes {
			actual := map[string]string{excludedTagName + "/something": "some-value"}
			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, map[string]string{}, actual)
		}
	})

	t.Run("excludes empty tags", func(t *testing.T) {
		actual := map[string]string{"good-tag": ""}
		cleanTags(actual, maxWavefrontTags)
		assert.Equal(t, map[string]string{}, actual)
	})

	t.Run("tags with dash values", func(t *testing.T) {
		t.Run("removes tags with - value when over capacity", func(t *testing.T) {
			actual := map[string]string{"non-dash-tag": "1234-5", "dash-tag": "-"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"non-dash-tag": "1234-5"}, actual)
		})
		t.Run("doesn't remove tags with '-' value when not over capacity", func(t *testing.T) {
			actual := map[string]string{"non-dash-tag": "1234-5", "dash-tag": "-"}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"non-dash-tag": "1234-5", "dash-tag": "-"}, actual)
		})
	})

	t.Run("tags with slash values", func(t *testing.T) {
		t.Run("removes tags with '/' value when over capacity", func(t *testing.T) {
			actual := map[string]string{"non-slash-tag": "1234/5", "slash-tag": "/"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"non-slash-tag": "1234/5"}, actual)
		})
		t.Run("doesn't remove tags with '/' value when not over capacity", func(t *testing.T) {
			actual := map[string]string{"non-slash-tag": "1234/5", "slash-tag": "/"}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"non-slash-tag": "1234/5", "slash-tag": "/"}, actual)
		})
	})

	t.Run("tags with label.*beta* in them", func(t *testing.T) {
		assert.True(t, betaRegex.MatchString("label.my.beta.prod"))
		t.Run("removes tags with 'label.*beta*' in name when over capacity", func(t *testing.T) {
			actual := map[string]string{"label.env": "prod", "label.my.beta.prod": "prod-beta"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"label.env": "prod"}, actual)
		})
		t.Run("doesn't remove tags 'label.*beta*' in name not over capacity", func(t *testing.T) {
			actual := map[string]string{"label.env": "prod", "label.my.beta.prod": "prod-beta"}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"label.env": "prod", "label.my.beta.prod": "prod-beta"}, actual)
		})
	})

	t.Run("de-duplicates tag values >= min dedupe value length characters when over capacity", func(t *testing.T) {
		tagGreaterThanMinLen := "some.hostname"
		assert.True(t, len(tagGreaterThanMinLen) >= minDedupeTagValueLen)

		tagEqualMinLen := "host1"
		assert.True(t, len(tagEqualMinLen) == minDedupeTagValueLen)

		tagLessThanMinLen := "host"
		assert.True(t, len(tagLessThanMinLen) < minDedupeTagValueLen)

		t.Run("when the tag names are different lengths", func(t *testing.T) {
			actual := map[string]string{"long-tag-name": tagGreaterThanMinLen, "shrt-tg": tagGreaterThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"shrt-tg": tagGreaterThanMinLen}, actual)
		})

		t.Run("when the tag names of the same length", func(t *testing.T) {
			actual := map[string]string{"dup2": tagGreaterThanMinLen, "dup1": tagGreaterThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"dup1": tagGreaterThanMinLen}, actual)
		})

		t.Run("when the duplicated values are < min len characters", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}, actual)
		})

		t.Run("when the duplicated values are equal min len characters", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagEqualMinLen, "b-tag": tagEqualMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"a-tag": tagEqualMinLen}, actual)
		})

		t.Run("when under the max capacity", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}, actual)
		})
	})
}
