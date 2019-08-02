package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/pflag"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/options"
	kubeFlag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
)

var (
	version string
	commit  string
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	opt := options.NewCollectorRunOptions()
	opt.AddFlags(pflag.CommandLine)
	kubeFlag.InitFlags()

	if opt.Version {
		fmt.Println(fmt.Sprintf("version: %s\ncommit: %s", version, commit))
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	//TODO: fix this

	//sinkManager := createAndInitSinksOrDie(opt.Sinks, opt.SinkExportDataTimeout)
	//
	//man, err := manager.NewFlushManager(nil, sinkManager, opt.MetricResolution)
	//if err != nil {
	//	log.Fatalf("Failed to create main manager: %v", err)
	//}
	//man.Start()
	waitForStop()
}

//func createAndInitSinksOrDie(sinkAddresses flags.Uris, sinkExportDataTimeout time.Duration) metrics.DataSink {
//	sinksFactory := sinks.NewSinkFactory()
//	sinkList := sinksFactory.BuildAll(sinkAddresses)
//
//	for _, sink := range sinkList {
//		log.Infof("Starting with %s", sink.Name())
//	}
//	sinkManager, err := sinks.NewDataSinkManager(sinkList, sinkExportDataTimeout, sinks.DefaultSinkStopTimeout)
//	if err != nil {
//		log.Fatalf("Failed to create sink manager: %v", err)
//	}
//	return sinkManager
//}

func waitForStop() {
	select {}
}
