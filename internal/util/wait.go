package util

import (
	"k8s.io/klog/v2"
	"net/http"
	"runtime"
	"time"
)

var (
	// ReallyCrash controls the behavior of HandleCrash and now defaults
	// true. It's still exposed so components can optionally set to false
	// to restore prior behavior.
	ReallyCrash = true

	// NeverStop may be passed to make it never stop
	NeverStop <-chan struct{} = make(chan struct{})

	// PanicHandlers is a list of functions which will be invoked when a panic happens.
	PanicHandlers = []func(interface{}){logPanic}
)

// Retry makes the function run infinitely after certain time period
func Retry(f func(), duration time.Duration, stopCh <-chan struct{}) {
	t := time.NewTicker(duration)

	for {
		<-t.C

		select {
		case <-stopCh:
			return
		default:
		}

		func() {
			defer HandleCrash()
			f()
		}()
	}
}

func HandleCrash(additionalHandlers ...func(interface{})) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
		if ReallyCrash {
			// Actually proceed to panic.
			panic(r)
		}
	}
}

// logPanic logs the caller tree when a panic occurs (except in the special case of http.ErrAbortHandler).
func logPanic(r interface{}) {
	if r == http.ErrAbortHandler {
		// honor the http.ErrAbortHandler sentinel panic value:
		//   ErrAbortHandler is a sentinel panic value to abort a handler.
		//   While any panic from ServeHTTP aborts the response to the client,
		//   panicking with ErrAbortHandler also suppresses logging of a stack trace to the server's error log.
		return
	}

	// Same as stdlib http server code. Manually allocate stack trace buffer size
	// to prevent excessively large logs
	const size = 64 << 10
	stacktrace := make([]byte, size)
	stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
	if _, ok := r.(string); ok {
		klog.Errorf("Observed a panic: %s\n%s", r, stacktrace)
	} else {
		klog.Errorf("Observed a panic: %#v (%v)\n%s", r, r, stacktrace)
	}
}
