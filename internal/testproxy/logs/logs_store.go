package logs

import (
	"encoding/json"
	"sync"
)

// Storing MissingTags as a map instead of a list for ease of lookup and adding
type LogStore struct {
	HasValidFormat bool                   `json:"hasValidFormat"`
	HasValidTags   bool                   `json:"hasValidTags"`
	MissingTags    []string               `json:"missingTags"`
	missingTagsMap map[string]interface{} `json:"-"`
	mu             *sync.Mutex            `json:"-"`
}

func NewLogStore() *LogStore {
	return &LogStore{
		HasValidFormat: false,
		missingTagsMap: make(map[string]interface{}),
		mu:             &sync.Mutex{},
	}
}

func (l *LogStore) SetReceivedWithValidFormat(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidFormat == value {
		return
	}

	l.HasValidFormat = value
}

func (l *LogStore) SetReceivedWithValidTags(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidTags == value {
		return
	}

	l.HasValidTags = value
}

func (l *LogStore) AddMissingTag(value string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.missingTagsMap[value]; ok {
		return
	}

	l.missingTagsMap[value] = ""
}

func (l *LogStore) ToJSON() (output []byte, err error) {
	l.updateMissingTagsForMarshalling()

	return json.Marshal(l)
}

func (l *LogStore) updateMissingTagsForMarshalling() {
	var missingTags []string

	l.mu.Lock()
	defer l.mu.Unlock()
	for k := range l.missingTagsMap {
		missingTags = append(missingTags, k)
	}

	l.MissingTags = missingTags
}
