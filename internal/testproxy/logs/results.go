package logs

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Results struct {
	// bool has no unset state
	HasValidFormat int `json:"hasValidFormat"`
	HasValidTags   int `json:"hasValidTags"`

	// TODO build this up over time, don't set every time
	missingExpectedTagsMap map[string]interface{} `json:"-"`
	MissingExpectedTags    []string               `json:"missingExpectedTags"`
	MissingExpectedCount   int                    `json:"missingExpectedCount"`

	emptyExpectedTagsMap map[string]interface{} `json:"-"`
	EmptyExpectedTags    []string               `json:"emptyExpectedTags"`
	EmptyExpectedCount   int                    `json:"emptyExpectedCount"`

	// TODO expect that we've never had a denied or missed an expected
	UnexpectedAllowedLogs  []interface{} `json:"unexpectedAllowedLogs"`
	UnexpectedAllowedCount int           `json:"unexpectedAllowedCount"`

	unexpectedDeniedTagsMap map[string]interface{} `json:"-"`
	UnexpectedDeniedTags    []string               `json:"unexpectedDeniedTags"`
	UnexpectedDeniedCount   int                    `json:"unexpectedDeniedCount"`

	ReceivedLogsCount int `json:"-"` // TODO export JSON and rebuild
	mu                *sync.Mutex
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
	if len(missingTags) > 0 {
		l.MissingExpectedCount++
	}

	for tag := range missingTags {
		l.missingExpectedTagsMap[tag] = nil
	}
}

func (l *Results) AddEmptyExpectedTags(emptyTags map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(emptyTags) > 0 {
		l.EmptyExpectedCount++
	}

	for tag := range emptyTags {
		l.emptyExpectedTagsMap[tag] = nil
	}
}

func (l *Results) AddUnexpectedAllowedLogs(unexpectedLogs []interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(unexpectedLogs) > 0 {
		l.UnexpectedAllowedCount++
	}

	l.UnexpectedAllowedLogs = append(l.UnexpectedAllowedLogs, unexpectedLogs...)
}

func (l *Results) AddUnexpectedDeniedTags(unexpectedDeniedTags map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(unexpectedDeniedTags) > 0 {
		l.UnexpectedDeniedCount++
	}

	for k, v := range unexpectedDeniedTags {
		tagKey := fmt.Sprintf("%s:%s", k, v)
		l.unexpectedDeniedTagsMap[tagKey] = nil
	}
}

func (l *Results) IncrementReceivedLogsCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.ReceivedLogsCount++
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
