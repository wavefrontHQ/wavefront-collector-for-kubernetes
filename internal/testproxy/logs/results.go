package logs

import (
	"encoding/json"
	"sync"
)

type Results struct {
	HasValidFormat       bool `json:"hasValidFormat"`
	HasValidTags         bool `json:"hasValidTags"`
	HasValidExpectedTags bool `json:"hasValidExpectedTags"`
	HasValidAllowedTags  bool `json:"hasValidAllowedTags"`
	HasValidDeniedTags   bool `json:"hasValidDeniedTags"`

	MissingExpectedTags []string `json:"missingExpectedTags"`
	EmptyExpectedTags   []string `json:"emptyExpectedTags"`

	UnexpectedAllowedLogs []interface{}          `json:"unexpectedAllowedLogs"`
	UnexpectedDeniedTags  map[string]interface{} `json:"unexpectedDeniedTags"`

	HasReceivedLogs bool `json:"-"`
	mu              *sync.Mutex
}

func NewLogStore() *Results {
	return &Results{
		mu: &sync.Mutex{},
	}
}

func (l *Results) SetHasValidFormat(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasValidFormat = value
}

func (l *Results) SetHasValidTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasValidTags = value
}

func (l *Results) SetHasValidExpectedTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasValidExpectedTags = value
}

func (l *Results) SetMissingExpectedTags(value []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.MissingExpectedTags = value
}

func (l *Results) SetEmptyExpectedTags(value []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.EmptyExpectedTags = value
}

func (l *Results) SetHasValidAllowedTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasValidAllowedTags = value
}

func (l *Results) SetUnexpectedAllowedLogs(value []interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedAllowedLogs = value
}

func (l *Results) SetHasValidDeniedTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasValidDeniedTags = value
}

func (l *Results) SetUnexpectedDeniedTags(value map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedDeniedTags = value
}

func (l *Results) SetHasReceivedLogs(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasReceivedLogs = value
}

func (l *Results) ToJSON() (output []byte, err error) {
	return json.Marshal(l)
}
