package logs

import (
	"encoding/json"
	"sync"
)

type LogStore struct {
	ReceivedWithValidFormat bool        `json:"receivedWithValidFormat"`
	mu                      *sync.Mutex `json:"-"`
}

func NewLogStore() *LogStore {
	return &LogStore{
		ReceivedWithValidFormat: false,
		mu:                      &sync.Mutex{},
	}
}

func (l *LogStore) SetReceivedWithValidFormat(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ReceivedWithValidFormat == value {
		return
	}

	l.ReceivedWithValidFormat = value
}

func (l *LogStore) ToJSON() (output []byte, err error) {
	return json.Marshal(l)
}
