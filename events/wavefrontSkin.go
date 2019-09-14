package events

import (
	"os"

	"github.com/wavefronthq/wavefront-sdk-go/event"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	v1 "k8s.io/api/core/v1"
)

type WavefrontSkin struct {
	sender senders.Sender
}

func NewWavefrontSkin() EventSinkInterface {
	directCfg := &senders.DirectConfiguration{
		Server: "https://nimba.wavefront.com",
		Token:  "6490a634-ca7d-47c1-bb04-4629f53fc98b",
	}

	sender, err := senders.NewDirectSender(directCfg)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	log.Print("sender ready")

	return &WavefrontSkin{sender: sender}
}

// UpdateEvents implements the EventSinkInterface
func (wf *WavefrontSkin) UpdateEvents(function string, eNew *v1.Event, eOld *v1.Event) {
	// new, err = json.Marshal(eNew)
	// if err != nil {
	// 	log.Warningf("Failed to json serialize event: %v", err)
	// }

	// log.WithField("new", string(new)).Info("UpdateEvents")

	ns := eNew.InvolvedObject.Namespace
	if len(ns) == 0 {
		ns = "default"
	}

	tags := []string{
		ns,
		eNew.InvolvedObject.Kind,
		eNew.Reason,
		function,
	}

	eType := eNew.Type
	if len(eType) == 0 {
		eType = "Normal"
	}

	wf.sender.SendEvent(
		eNew.Message,
		eNew.LastTimestamp.Unix(),
		eNew.LastTimestamp.Unix()+1, // TODO: remove
		eNew.InvolvedObject.Name,
		tags,
		event.Type(eType),
	)
}
