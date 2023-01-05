package logs

import (
	"encoding/json"
	"sync"
)

// allowlist
// check each log line to make sure that at least one allow list tag/value pair is in its tag list
// if it doesn't have any from the allow list, that's an invalid log message that should have been dropped by fluentd
// and means we have configured fluentd wrong. This should be an integration test failure as a result.

// denylist
// check each log line to make sure that none of the deny list tag/value pairs exist in it.
// If any do, that's an invalid log message that should have been dropped by fluentd
// and means we have configured fluentd wrong. This should also be an integration test failure as a result.

// assumptions before checking allowlist/denylist
// 1. We are receiving data into the test proxy
// 2. It is not invalid data for the endpoint
// 3. The format of the logs are valid
// 4. There are tags and they are the expected minimum ones

/*
  if never received data

	{"receivedLogs": false}

  if received data in wrong format:

	{
		"receivedLogs": true,
		"validFormat": false
	}


  if received data with missing expected tags:

	{
		"receivedLogs": true,
		"validFormat": true,
		"missingTags": ["some-missing", "another-missing"]
	}

  if received data with expected tags, but no tags match the allowlist :

	{
		"receivedLogs": true,
		"validFormat": true,
		"missingExpectedTags": [],
        "receivedLogsMissingAllowListTags": 10
	}

  if received data with expected tags and some tag from the allowlist, but has tags from the denylist :

	{
		"receivedLogs": true,
		"validFormat": true,
		"missingTags": [],
        "receivedMissingAllowListTags": 0,
        "receivedWithDenyListTags": true
	}

  if received data with expected tags, tags from the allowlist and no tags from the denylist


	{
		"receivedLogs": true,
		"validFormat": true,
		"missingTags": [],
        "receivedMissingAllowListTags": 0,
        "receivedWithDenyListTags": false
	}
*/

type Results struct {
	HasValidFormat             bool        `json:"hasValidFormat"`
	HasValidTags               bool        `json:"hasValidTags"`
	MissingExpectedTags        []string    `json:"missingExpectedTags"`
	AllowListFilterMissedCount int         `json:"allowListFilterMissedCount"`
	DenyListFilterMissedCount  int         `json:"denyListFilterMissedCount"`
	HasReceivedLogs            bool        `json:"-"`
	mu                         *sync.Mutex `json:"-"`
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

func (l *Results) SetMissingTags(value []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.MissingExpectedTags = value
}

func (l *Results) SetHasReceivedLogs(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.HasReceivedLogs = value
}

func (l *Results) ToJSON() (output []byte, err error) {
	return json.Marshal(l)
}
