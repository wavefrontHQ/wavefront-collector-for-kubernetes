package cadvisor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
)

func TestTransformMetricsPermissionError(t *testing.T) {
	t.Run("does not transform errors that are not a prometheus.HTTPError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		actualErr := TransformPermissionError(expectedErr)
		assert.Equal(t, expectedErr.Error(), actualErr.Error())
	})

	t.Run("transforms a prometheus.HTTPError when its status code is 403", func(t *testing.T) {
		inputErr := &prometheus.HTTPError{MetricsURL: "http://localhost", Status: "403 Forbidden", StatusCode: 403}
		actualErr := TransformPermissionError(inputErr)
		assert.Contains(t, actualErr.Error(), "missing nodes/metrics permission in the collector's cluster role")
	})

	t.Run("transforms a prometheus.HTTPError when its status code is 401", func(t *testing.T) {
		inputError := &prometheus.HTTPError{MetricsURL: "http://localhost", Status: "401 Unauthorized", StatusCode: 401}
		actualErr := TransformPermissionError(inputError)
		assert.Contains(t, actualErr.Error(), "missing nodes/metrics permission in the collector's cluster role")
	})

	t.Run("does not transform a prometheus.HTTPError when its status code is not 401 or 403", func(t *testing.T) {
		expectedErr := &prometheus.HTTPError{MetricsURL: "http://localhost", Status: "500 Internal Error", StatusCode: 500}
		actualErr := TransformPermissionError(expectedErr)
		assert.Equal(t, expectedErr.Error(), actualErr.Error())
	})

	t.Run("does not transform a nil prometheus.HTTPError error", func(t *testing.T) {
		var expectedErr *prometheus.HTTPError
		actualErr := TransformPermissionError(expectedErr)
		assert.Nil(t, actualErr)
	})

	t.Run("does not transform a nil error", func(t *testing.T) {
		var expectedErr error
		actualErr := TransformPermissionError(expectedErr)
		assert.Nil(t, actualErr)
	})
}
