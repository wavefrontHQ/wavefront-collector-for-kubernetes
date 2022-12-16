package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/test-proxy/handlers"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/test-proxy/metrics"
	"net"
	"net/http"
	"os"
)

var proxyAddr = ":7777"
var controlAddr = ":8888"
var runMode = "metrics"
var logLevel = log.InfoLevel.String()

func init() {
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, "host and port for the test \"wavefront proxy\" to listen on")
	flag.StringVar(&controlAddr, "control", controlAddr, "host and port for the http control server to listen on")
	flag.StringVar(&runMode, "mode", runMode, "which mode to run in. Valid options are \"metrics\", and \"logs\"")
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

	// Set log output to file to prevent our logging component from picking up stdout/stderr logs
	// and sending them back to us over and over.
	file, err := os.OpenFile("test-proxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could not create log output file: ", err)
		os.Exit(1)
	}

	log.SetOutput(file)

	store := metrics.NewMetricStore()

	switch runMode {
	case "metrics":
		go serveMetrics(store)
	case "logs":
		go serveLogs()
	default:
		log.Error("\"mode\" flag must be set to: \"metrics\", or \"logs\"")
		os.Exit(1)
	}

	// Blocking call to start up the control server
	serveControl(store)
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

func serveLogs() {
	logsServeMux := http.NewServeMux()
	logsServeMux.HandleFunc("/logs/json_array", handlers.LogJsonArrayHandler())
	logsServeMux.HandleFunc("/logs/json_lines", handlers.LogJsonLinesHandler())

	log.Infof("http logs server listening on %s", proxyAddr)
	if err := http.ListenAndServe(proxyAddr, logsServeMux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func serveControl(store *metrics.MetricStore) {
	controlServeMux := http.NewServeMux()

	controlServeMux.HandleFunc("/metrics", handlers.DumpMetricsHandler(store))
	controlServeMux.HandleFunc("/metrics/diff", handlers.DiffMetricsHandler(store))
	// Based on logs already sent, perform checks on store logs
	// Start by supporting POST parameter expected_log_format
	controlServeMux.HandleFunc("/logs/assert", handlers.LogAssertionHandler())
	// NOTE: these handler functions attach to the control HTTP server, NOT the TCP server that actually receives data

	log.Infof("http control server listening on %s", controlAddr)
	if err := http.ListenAndServe(controlAddr, controlServeMux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
