package events

import (
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
)

type EventSink interface {
	ExportEvent(*Event)
}

type Event struct {
	Message string
	Ts      time.Time
	Host    string
	Tags    map[string]string
	Options []event.Option
}
