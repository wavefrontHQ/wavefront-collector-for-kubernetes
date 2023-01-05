package logs

import (
	"encoding/json"
	"sync"
)

type LogStore struct {
	HasValidFormat       bool              `json:"hasValidFormat"`
	HasValidTags         bool              `json:"hasValidTags"`
	MissingExpectedTags  []string          `json:"missingExpectedTags"`
	MissingAllowListTags map[string]string `json:"missingAllowListTags"`
	ExtraDenyListTags    map[string]string `json:"extraDenyListTags"`
	HasReceivedLogs      bool              `json:"-"`
	mu                   *sync.Mutex       `json:"-"`
}

func NewLogStore() *LogStore {
	return &LogStore{
		MissingAllowListTags: make(map[string]string),
		ExtraDenyListTags:    make(map[string]string),
		mu:                   &sync.Mutex{},
	}
}

func (l *LogStore) SetHasValidFormat(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidFormat == value {
		return
	}

	l.HasValidFormat = value
}

func (l *LogStore) SetHasValidTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidTags == value {
		return
	}

	l.HasValidTags = value
}

func (l *LogStore) SetMissingTags(value []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.MissingExpectedTags = value
}

func (l *LogStore) SetMissingAllowListTags(value map[string]string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.MissingAllowListTags = value
}

func (l *LogStore) SetExtraDenyListTags(value map[string]string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ExtraDenyListTags = value
}

func (l *LogStore) SetHasReceivedLogs(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.HasReceivedLogs = value
}

func (l *LogStore) ToJSON() (output []byte, err error) {
	return json.Marshal(l)
}
