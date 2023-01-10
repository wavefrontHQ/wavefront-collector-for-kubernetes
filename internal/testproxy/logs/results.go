package logs

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Results struct {
	HasReceivedLogs bool `json:"hasReceivedLogs"`
	HasValidFormat  int  `json:"hasValidFormat"`
	HasValidTags    int  `json:"hasValidTags"`

	missingExpectedTagsMap map[string]interface{} `json:"-"`
	MissingExpectedTags    []string               `json:"missingExpectedTags"`

	emptyExpectedTagsMap map[string]interface{} `json:"-"`
	EmptyExpectedTags    []string               `json:"emptyExpectedTags"`

	UnexpectedAllowedLogs []interface{} `json:"unexpectedAllowedLogs"`

	unexpectedDeniedTagsMap map[string]interface{} `json:"-"`
	UnexpectedDeniedTags    []string               `json:"unexpectedDeniedTags"`
	mu                      *sync.Mutex
}

func NewLogStore() *Results {
	return &Results{
		HasValidFormat:          -1,
		HasValidTags:            -1,
		missingExpectedTagsMap:  make(map[string]interface{}),
		emptyExpectedTagsMap:    make(map[string]interface{}),
		unexpectedDeniedTagsMap: make(map[string]interface{}),
		mu:                      &sync.Mutex{},
	}
}

func (l *Results) SetHasReceivedLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.HasReceivedLogs = true
}

func (l *Results) SetHasValidFormat(isValid bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidFormat == 0 {
		return
	}

	if isValid {
		l.HasValidFormat = 1
	} else {
		l.HasValidFormat = 0
	}
}

func (l *Results) SetHasValidTags(isValid bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.HasValidTags == 0 {
		return
	}

	if isValid {
		l.HasValidTags = 1
	} else {
		l.HasValidTags = 0
	}
}

func (l *Results) AddMissingExpectedTags(missingTags map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for tag := range missingTags {
		l.missingExpectedTagsMap[tag] = nil
	}
}

func (l *Results) AddEmptyExpectedTags(emptyTags map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for tag := range emptyTags {
		l.emptyExpectedTagsMap[tag] = nil
	}
}

func (l *Results) AddUnexpectedAllowedLogs(unexpectedLogs []interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedAllowedLogs = append(l.UnexpectedAllowedLogs, unexpectedLogs...)
}

func (l *Results) AddUnexpectedDeniedTags(unexpectedDeniedTags map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for k, v := range unexpectedDeniedTags {
		tagKey := fmt.Sprintf("%s:%s", k, v)
		l.unexpectedDeniedTagsMap[tagKey] = nil
	}
}

func (l *Results) ToJSON() (output []byte, err error) {
	l.MissingExpectedTags = mapKeysToSlice(l.missingExpectedTagsMap)
	l.EmptyExpectedTags = mapKeysToSlice(l.emptyExpectedTagsMap)
	l.UnexpectedDeniedTags = mapKeysToSlice(l.unexpectedDeniedTagsMap)

	return json.Marshal(l)
}

func mapKeysToSlice(m map[string]interface{}) []string {
	var output []string
	for k := range m {
		output = append(output, k)
	}

	return output
}
