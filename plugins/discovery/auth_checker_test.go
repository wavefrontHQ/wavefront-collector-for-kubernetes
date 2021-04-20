package discovery_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/discovery"
	v1 "k8s.io/api/authorization/v1"
)

func TestAuthChecker(t *testing.T) {
	t.Run("auth checker caching", func(t *testing.T) {
		t.Run("initializes", func(t *testing.T) {
			spy := &AccessSpy{allowed: true}
			checker := discovery.NewAuthChecker(spy, "namespace", time.Hour, time.Hour)

			assert.True(t, checker.CanListSecrets())
		})

		t.Run("caches access", func(t *testing.T) {
			spy := &AccessSpy{allowed: true}
			checker := discovery.NewAuthChecker(spy, "namespace", time.Hour, time.Hour)

			assert.True(t, checker.CanListSecrets())
			assert.True(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(spy.parameters), "Premature repeat api call")
		})

		t.Run("refreshes access", func(t *testing.T) {
			spy := &AccessSpy{allowed: true}
			checker := discovery.NewAuthChecker(spy, "namespace", 1*time.Nanosecond, time.Hour)

			assert.True(t, checker.CanListSecrets())
			assert.True(t, checker.CanListSecrets())
			assert.Equal(t, 2, len(spy.parameters), "Failed to Refresh")
		})
	})

	t.Run("auth checker logging", func(t *testing.T) {
		t.Run("No log with access", func(t *testing.T) {
			spy := &AccessSpy{allowed: true}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, time.Hour, logSpy.infof)

			assert.True(t, checker.CanListSecrets())
			assert.Equal(t, 0, len(logSpy.messages))
		})
		t.Run("Log with no access", func(t *testing.T) {
			spy := &AccessSpy{allowed: false}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, time.Hour, logSpy.infof)

			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(logSpy.messages))
		})
		t.Run("Only log once with no access", func(t *testing.T) {
			spy := &AccessSpy{allowed: false}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, time.Hour, logSpy.infof)

			assert.False(t, checker.CanListSecrets())
			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(logSpy.messages))
		})
		t.Run("Log again after interval expires", func(t *testing.T) {
			spy := &AccessSpy{allowed: false}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, 1*time.Nanosecond, logSpy.infof)

			assert.False(t, checker.CanListSecrets())
			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 2, len(logSpy.messages))
		})
		t.Run("Log lost access", func(t *testing.T) {
			spy := &AccessSpy{allowed: true}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, 1*time.Hour, logSpy.infof)

			assert.True(t, checker.CanListSecrets())
			assert.Equal(t, 0, len(logSpy.messages))

			spy.allowed = false
			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(logSpy.messages))
		})
		t.Run("Log toggle access", func(t *testing.T) {
			spy := &AccessSpy{allowed: false}
			logSpy := &LogSpy{}
			checker := discovery.TestAuthChecker(spy, "namespace", 1*time.Nanosecond, 1*time.Hour, logSpy.infof)

			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(logSpy.messages))

			spy.allowed = true
			assert.True(t, checker.CanListSecrets())
			assert.Equal(t, 1, len(logSpy.messages))

			spy.allowed = false
			assert.False(t, checker.CanListSecrets())
			assert.Equal(t, 2, len(logSpy.messages))
		})

	})
}

type AccessSpy struct {
	parameters []*v1.ResourceAttributes
	allowed    bool
}

func (spy *AccessSpy) Create(sar *v1.SelfSubjectAccessReview) (result *v1.SelfSubjectAccessReview, err error) {
	spy.parameters = append(spy.parameters, sar.Spec.ResourceAttributes)
	return &v1.SelfSubjectAccessReview{
		Status: v1.SubjectAccessReviewStatus{
			Allowed: spy.allowed,
		},
	}, nil
}

type LogSpy struct {
	messages []string
}

func (spy *LogSpy) infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	spy.messages = append(spy.messages, msg)
}
