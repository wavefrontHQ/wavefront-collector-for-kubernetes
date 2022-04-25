package util

import (
	"time"
)

var (
	// NeverStop may be passed to make it never stop
	NeverStop <-chan struct{} = make(chan struct{})
)

// TODO: this needs a test! Tricky...
// Retry makes the function run infinitely after certain time period
func Retry(f func(), duration time.Duration, stopCh <-chan struct{}) {
	t := time.NewTicker(duration)

	for {
		select {
		case <-stopCh:
			return
		case <-t.C:
		}

		select {
		case <-stopCh:
			return
		default:
		}

		f()
	}
}
