package logs

import (
	"encoding/json"
	"sync"
)

type Results struct {
	HasValidFormat int `json:"hasValidFormat"`
	HasValidTags   int `json:"hasValidTags"`

	MissingExpectedTagsMap   map[string]interface{} `json:"-"`
	MissingExpectedTags      []string               `json:"missingExpectedTags"`
	MissingExpectedTagsCount int                    `json:"missingExpectedTagsCount"`

	EmptyExpectedTagsMap   map[string]interface{} `json:"-"`
	EmptyExpectedTags      []string               `json:"emptyExpectedTags"`
	EmptyExpectedTagsCount int                    `json:"emptyExpectedTagsCount"`

	UnexpectedAllowedLogs      []interface{} `json:"unexpectedAllowedLogs"`
	UnexpectedAllowedLogsCount int           `json:"unexpectedAllowedLogsCount"`

	UnexpectedDeniedTagsMap   map[string]interface{} `json:"-"`
	UnexpectedDeniedTags      []string               `json:"unexpectedDeniedTags"`
	UnexpectedDeniedTagsCount int                    `json:"unexpectedDeniedTagsCount"`

	ReceivedLogCount int `json:"receivedLogCount"`
	mu               *sync.Mutex
}

// When using this, only modify struct fields through the receiver methods
//
// They need to be exposed for converting to JSON and we use
// the fact that they are exposed for ease of testing.
func NewLogResults() *Results {
	return &Results{
		HasValidFormat:          -1,
		HasValidTags:            -1,
		MissingExpectedTagsMap:  make(map[string]interface{}),
		EmptyExpectedTagsMap:    make(map[string]interface{}),
		UnexpectedDeniedTagsMap: make(map[string]interface{}),
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

func (l *Results) AddMissingExpectedTag(missingTag string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.MissingExpectedTagsMap[missingTag] = nil
}

func (l *Results) AddEmptyExpectedTag(emptyTag string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.EmptyExpectedTagsMap[emptyTag] = nil
}

func (l *Results) AddUnexpectedAllowedLog(unexpectedLog interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedAllowedLogs = append(l.UnexpectedAllowedLogs, unexpectedLog)
}

func (l *Results) AddUnexpectedDeniedTag(unexpectedDeniedTag string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedDeniedTagsMap[unexpectedDeniedTag] = nil
}

func (l *Results) IncrementMissingExpectedTagsCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.MissingExpectedTagsCount++
}

func (l *Results) IncrementEmptyExpectedTagsCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.EmptyExpectedTagsCount++
}

func (l *Results) IncrementUnexpectedAllowedLogsCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedAllowedLogsCount++
}

func (l *Results) IncrementUnexpectedDeniedTagsCount() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.UnexpectedDeniedTagsCount++
}

func (l *Results) AddReceivedCount(count int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.ReceivedLogCount += count
}

func (l *Results) ToJSON() (output []byte, err error) {
	l.MissingExpectedTags = mapKeysToSlice(l.MissingExpectedTagsMap)
	l.EmptyExpectedTags = mapKeysToSlice(l.EmptyExpectedTagsMap)
	l.UnexpectedDeniedTags = mapKeysToSlice(l.UnexpectedDeniedTagsMap)

	return json.Marshal(l)
}

func mapKeysToSlice(m map[string]interface{}) []string {
	var output []string
	for k := range m {
		output = append(output, k)
	}

	return output
}
