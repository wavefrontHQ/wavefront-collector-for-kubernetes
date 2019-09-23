package events

import (
	"fmt"
	"os"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-sdk-go/event"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	v1 "k8s.io/api/core/v1"
)

type WavefrontSkin struct {
	sender      senders.Sender
	clusterName string
}

func NewWavefrontSkin(wf *configuration.WavefrontSinkConfig, clusterName string) EventSinkInterface {
	if len(wf.Server) == 0 || len(wf.Token) == 0 {
		log.Fatal("Invalid EventSink configuration `Server` and `Token` are required")
	}

	directCfg := &senders.DirectConfiguration{
		Server: wf.Server,
		Token:  wf.Token,
	}

	sender, err := senders.NewDirectSender(directCfg)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	log.Print("sender ready")

	return &WavefrontSkin{sender: sender, clusterName: clusterName}
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
		fmt.Sprintf("ns: %s", ns),
		fmt.Sprintf("k8s: %s", wf.clusterName),
		fmt.Sprintf("Kind: %s", eNew.InvolvedObject.Kind),
		fmt.Sprintf("Reason: %s", eNew.Reason),
		fmt.Sprintf("cmt: %s", eNew.Source.Component),
	}

	eType := eNew.Type
	if len(eType) == 0 {
		eType = "Normal"
	}

	wf.sender.SendEvent(
		eNew.Message,
		eNew.LastTimestamp.Unix(), 0,
		eNew.Source.Host,
		tags,
		event.Type(eType),
	)
}
