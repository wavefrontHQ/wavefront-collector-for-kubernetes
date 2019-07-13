package util

import (
	"os"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
)

type FileListener interface {
	Changed(file string)
}

type ConfigHandler interface {
	Handle(cfg interface{})
}

type FileWatcher interface {
	Watch()
	Stop()
}

type fileWatcher struct {
	file     string
	listener FileListener
	delay    time.Duration
	modTime  time.Time
	stopCh   chan struct{}
}

func NewFileWatcher(file string, listener FileListener, initialDelay time.Duration) FileWatcher {
	return &fileWatcher{
		file:     file,
		listener: listener,
		delay:    initialDelay,
		stopCh:   make(chan struct{}),
	}
}

// listens for changes to a given file every minute
func (fw *fileWatcher) Watch() {
	fw.stopCh = make(chan struct{})
	initial := true
	go wait.Until(func() {
		if initial {
			time.Sleep(fw.delay)
		}
		fileInfo, err := os.Stat(fw.file)
		if err != nil {
			glog.Errorf("error retrieving file stats: %v", err)
			return
		}

		if fileInfo.ModTime().After(fw.modTime) {
			fw.modTime = fileInfo.ModTime()
			if !initial {
				fw.listener.Changed(fw.file)
			} else {
				initial = false
			}
		}
	}, 1*time.Minute, fw.stopCh)
}

func (fw *fileWatcher) Stop() {
	close(fw.stopCh)
}
