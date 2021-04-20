package discovery

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/authorization/v1"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type AuthChecker struct {
	accessGetter authv1.SelfSubjectAccessReviewInterface
	namespace    string

	hasAccess bool

	refreshInterval time.Duration
	reportInterval  time.Duration
	lastChecked     time.Time
	lastLogged      time.Time
	logger          func(format string, args ...interface{})
}

func NewAuthChecker(accessGetter authv1.SelfSubjectAccessReviewInterface, namespace string, refreshInterval time.Duration, reportInterval time.Duration) *AuthChecker {
	return TestAuthChecker(accessGetter, namespace, refreshInterval, reportInterval, log.Infof)
}

func TestAuthChecker(accessGetter authv1.SelfSubjectAccessReviewInterface, namespace string, refreshInterval time.Duration, reportInterval time.Duration, logger func(format string, args ...interface{})) *AuthChecker {
	checker := &AuthChecker{
		accessGetter:    accessGetter,
		namespace:       namespace,
		refreshInterval: refreshInterval,
		reportInterval:  reportInterval,
		logger:          logger,
	}

	return checker
}

func (checker *AuthChecker) CanListSecrets() bool {
	checker.refreshAccess()
	return checker.hasAccess
}

func (checker *AuthChecker) refreshAccess() {
	if !checker.timeToRefresh() {
		return
	}
	checker.lastChecked = time.Now()

	checker.hasAccess = checker.canListSecretsAPI()
	checker.reportAccess()
}

func (checker *AuthChecker) timeToRefresh() bool {
	return time.Now().Sub(checker.lastChecked) > checker.refreshInterval
}

func (checker *AuthChecker) reportAccess() {
	if checker.hasAccess {
		return
	}
	if time.Now().Sub(checker.lastLogged) < checker.reportInterval {
		return
	}

	checker.lastLogged = time.Now()
	checker.logger("Secret Access Not Available for Configuration")
}

func (checker *AuthChecker) canListSecretsAPI() bool {
	sar := &v1.SelfSubjectAccessReview{
		Spec: v1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Namespace: checker.namespace,
				Verb:      "list",
				Resource:  "Secret",
			},
		},
	}
	review, err := checker.accessGetter.Create(sar)
	if err != nil {
		log.Errorf("Unable to check api access: %v", err)
		return false
	}
	return review.Status.Allowed
}
