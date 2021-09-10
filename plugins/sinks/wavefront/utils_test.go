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

	t.Run("de-duplicates tag values >= 10 characters when over capacity", func(t *testing.T) {
		t.Run("when the tag names are different lengths", func(t *testing.T) {
			actual := map[string]string{"long-tag-name": "some.hostname", "shrt-tg": "some.hostname"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"shrt-tg": "some.hostname"}, actual)
		})

		t.Run("when the tag names of the same length", func(t *testing.T) {
			actual := map[string]string{"dup2": "some.hostname", "dup1": "some.hostname"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"dup1": "some.hostname"}, actual)
		})

		t.Run("when the duplicated values are < 10 characters", func(t *testing.T) {
			actual := map[string]string{"a-tag": "123", "b-tag": "123"}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"a-tag": "123", "b-tag": "123"}, actual)
		})

		t.Run("when under the max capacity", func(t *testing.T) {
			actual := map[string]string{"a-tag": "123", "b-tag": "123"}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"a-tag": "123", "b-tag": "123"}, actual)
		})
	})
}
