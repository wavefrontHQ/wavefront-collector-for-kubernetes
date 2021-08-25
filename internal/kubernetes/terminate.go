package kubernetes

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

var testMode = false
var TerminationMessage string

func UseTerminateTestMode() {
	testMode = true
}

func Terminate(message string) {
	if testMode {
		TerminationMessage = message
	} else {
		err := ioutil.WriteFile("/dev/termination-log", []byte(message), 0644)
		if err != nil {
			log.Error(err)
		}
		log.Fatal(message)
	}
}
