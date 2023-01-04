package logs

import (
	"encoding/json"
	"sync"
)

type LogStore struct {
	HasValidFormat  bool        `json:"hasValidFormat"`
	HasValidTags    bool        `json:"hasValidTags"`
	MissingTags     []string    `json:"missingTags"`
	HasReceivedLogs bool        `json:"-"`
	mu              *sync.Mutex `json:"-"`
}

func NewLogStore() *LogStore {
	return &LogStore{
		mu: &sync.Mutex{},
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
	l.MissingTags = value
}

func (l *LogStore) SetHasReceivedLogs(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.HasReceivedLogs = value
}

func (l *LogStore) ToJSON() (output []byte, err error) {
	return json.Marshal(l)
}
