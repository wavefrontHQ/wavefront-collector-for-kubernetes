package main

import (
	"bufio"
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
)

var proxyAddr = ":7777"
var controlAddr = ":8888"

func init() {
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, "host and port for the test \"wavefront proxy\" to listen on")
	flag.StringVar(&controlAddr, "control", controlAddr, "host and port for the http control server to listen on")
}

func main() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	store := NewMetricStore()

	go func() {
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
			go HandleIncomingMetrics(store, conn)
		}
	}()

	http.HandleFunc("/metrics", DumpMetricsHandler(store))
	http.HandleFunc("/metrics/diff", DiffMetricsHandler(store)) // ?prefixes=prom-example,foo,bar&spec=names,values,timestamps,tagnames,tagvalues
	log.Infof("http control server listening on %s", controlAddr)
	if err := http.ListenAndServe(controlAddr, nil); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func HandleIncomingMetrics(store *MetricStore, conn net.Conn) {
	defer conn.Close()
	lines := bufio.NewScanner(conn)
	for lines.Scan() {
		if len(lines.Text()) == 0 {
			continue
		}
		metric, err := ParseMetric(lines.Text())
		if err != nil {
			log.Error(err.Error())
			log.Error(lines.Text())
			store.LogBadMetric(lines.Text())
			continue
		}
		log.Infof("%#v", metric)
		store.LogMetric(metric)
	}
	if err := lines.Err(); err != nil {
		log.Error(err.Error())
	}
	return
}

func DumpMetricsHandler(store *MetricStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Errorf("expected method %s but got %s", http.MethodGet, req.Method)
			return
		}
		badMetrics := store.BadMetrics()
		if len(badMetrics) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			err := json.NewEncoder(w).Encode(badMetrics)
			if err != nil {
				log.Error(err.Error())
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(store.Metrics())
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func DiffMetricsHandler(store *MetricStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Errorf("expected method %s but got %s", http.MethodGet, req.Method)
			return
		}
		badMetrics := store.BadMetrics()
		if len(badMetrics) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			err := json.NewEncoder(w).Encode(badMetrics)
			if err != nil {
				log.Error(err.Error())
			}
			return
		}
		var expectedMetrics []*Metric
		lines := bufio.NewScanner(req.Body)
		defer req.Body.Close()
		for lines.Scan() {
			if len(lines.Bytes()) == 0 {
				continue
			}
			var expectedMetric *Metric
			err := json.Unmarshal(lines.Bytes(), &expectedMetric)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(err.Error())
				if err != nil {
					log.Error(err.Error())
				}
				return
			}
			expectedMetrics = append(expectedMetrics, expectedMetric)
		}
		linesErr := lines.Err()
		if linesErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			ioErr := json.NewEncoder(w).Encode(linesErr.Error())
			if ioErr != nil {
				log.Error(ioErr.Error())
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		linesErr = json.NewEncoder(w).Encode(DiffMetrics(expectedMetrics, store.Metrics()))
		if linesErr != nil {
			log.Error(linesErr.Error())
		}
	}
}
