package util

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

var (
	// panicHandlers is a list of functions which will be invoked when a panic happens.
	panicHandlers = []func(interface{}){logPanic}
)

func HandleCrash(additionalHandlers ...func(interface{})) {
	if r := recover(); r != nil {
		for _, fn := range panicHandlers {
			fn(r)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
		panic(r)
	}
}

// logPanic logs the caller tree when a panic occurs (except in the special case of http.ErrAbortHandler).
func logPanic(r interface{}) {
	if r == http.ErrAbortHandler {
		//   honor the http.ErrAbortHandler sentinel panic value:
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

// errorHandlers is a list of functions which will be invoked when a nonreturnable
// error occurs.
var errorHandlers = []func(error){
	logError,
	(&rudimentaryErrorBackoff{
		lastErrorTime: time.Now(),
		minPeriod:     time.Millisecond,
	}).onError,
}

type rudimentaryErrorBackoff struct {
	minPeriod         time.Duration
	lastErrorTimeLock sync.Mutex
	lastErrorTime     time.Time
}

// onError will block if it is called more often than the embedded period time.
// This will prevent overly tight hot error loops.
func (r *rudimentaryErrorBackoff) onError(error) {
	r.lastErrorTimeLock.Lock()
	defer r.lastErrorTimeLock.Unlock()
	d := time.Since(r.lastErrorTime)
	if d < r.minPeriod {
		// If the time moves backwards for any reason, do nothing
		time.Sleep(r.minPeriod - d)
	}
	r.lastErrorTime = time.Now()
}

// HandleError handles log on a backoff loop
func HandleError(err error) {
	if err == nil {
		return
	}
	for _, fn := range errorHandlers {
		fn(err)
	}
}

// logError prints an error with the call stack of the location it was reported
func logError(err error) {
	klog.ErrorDepth(2, err)
}
