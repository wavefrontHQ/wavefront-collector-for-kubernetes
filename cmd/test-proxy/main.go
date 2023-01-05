package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/handlers"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/metrics"
)

var (
	proxyAddr   = ":7777"
	controlAddr = ":8888"
	runMode     = "metrics"
	logFilePath string
	logLevel    = log.InfoLevel.String()

	expectedTags = []string{"user_defined_tag",
		"service",
		"application",
		"source",
		"cluster",
		"timestamp",
		"pod_name",
		"container_name",
		"namespace_name",
		"pod_id",
		"container_id",
	}
	allowListFilteredTags = map[string]string{
		"namespace_name": "kube-system",
	}
	denyListFilteredTags = map[string]string{
		"container_name": "kube-apiserver",
	}
)

func init() {
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, "host and port for the test \"wavefront proxy\" to listen on")
	flag.StringVar(&controlAddr, "control", controlAddr, "host and port for the http control server to listen on")
	flag.StringVar(&runMode, "mode", runMode, "which mode to run in. Valid options are \"metrics\", and \"logs\"")
	flag.StringVar(&logFilePath, "logFilePath", logFilePath, "the full path to output logs to instead of using stdout")
	flag.StringVar(&logLevel, "logLevel", logLevel, "change log level. Default is \"info\", use \"debug\" for metric logging")
}

func main() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})
	if level, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	metricStore := metrics.NewMetricStore()
	logStore := logs.NewLogStore()

	if logFilePath != "" {
		// Set log output to file to prevent our logging component from picking up stdout/stderr logs
		// and sending them back to us over and over.
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Could not create log output file: ", err)
			os.Exit(1)
		}

		log.SetOutput(file)
	}

	switch runMode {
	case "metrics":
		go serveMetrics(metricStore)
	case "logs":
		go serveLogs(logStore)
	default:
		log.Error("\"mode\" flag must be set to: \"metrics\" or \"logs\"")
		os.Exit(1)
	}

	// Blocking call to start up the control server
	serveControl(metricStore, logStore)
}

func serveMetrics(store *metrics.MetricStore) {
	log.Infof("tcp metrics server listening on %s", proxyAddr)
	listener, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		go handlers.HandleIncomingMetrics(store, conn)
	}
}

func serveLogs(store *logs.LogStore) {
	logVerifier := logs.NewLogVerifier(expectedTags, allowListFilteredTags, denyListFilteredTags)

	logsServeMux := http.NewServeMux()
	logsServeMux.HandleFunc("/logs/json_array", handlers.LogJsonArrayHandler(logVerifier, store))
	logsServeMux.HandleFunc("/logs/json_lines", handlers.LogJsonLinesHandler(logVerifier, store))

	log.Infof("http logs server listening on %s", proxyAddr)
	if err := http.ListenAndServe(proxyAddr, logsServeMux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func serveControl(metricStore *metrics.MetricStore, logStore *logs.LogStore) {
	controlServeMux := http.NewServeMux()

	controlServeMux.HandleFunc("/metrics", handlers.DumpMetricsHandler(metricStore))
	controlServeMux.HandleFunc("/metrics/diff", handlers.DiffMetricsHandler(metricStore))
	// Based on logs already sent, perform checks on store logs
	// Start by supporting POST parameter expected_log_format
	controlServeMux.HandleFunc("/logs/assert", handlers.LogAssertionHandler(logStore))
	// NOTE: these handler functions attach to the control HTTP server, NOT the TCP server that actually receives data

	log.Infof("http control server listening on %s", controlAddr)
	if err := http.ListenAndServe(controlAddr, controlServeMux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
