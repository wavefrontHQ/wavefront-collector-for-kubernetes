package util

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"time"
)

// NeverStop may be passed to make it never stop
var NeverStop <-chan struct{} = make(chan struct{})

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
            defer runtime.HandleCrash()
            f()
        }()
	}
}
