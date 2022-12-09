package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var proxyAddr = ":7777"
var logsPort = ":9999"
var controlAddr = ":8888"
var logLevel = log.InfoLevel.String()
var logsPath = "/logs/json_array"

func init() {
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, "host and port for the test \"wavefront proxy\" to listen on")
	flag.StringVar(&logsPort, "logsPort", logsPort, "port for the logs server to listen on")
	flag.StringVar(&controlAddr, "control", controlAddr, "host and port for the http control server to listen on")
	flag.StringVar(&logsPath, "logsPath", logsPath, "the URL path for the test \"wavefront proxy\" to listen on for logging data.")
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
	log.SetOutput(os.Stdout)

	//store := NewMetricStore()

	//go ServeProxy(store)
	//http.HandleFunc("/metrics", DumpMetricsHandler(store))
	//http.HandleFunc("/metrics/diff", DiffMetricsHandler(store))
	log.Infof("logs port listening on %s", logsPort)
	http.HandleFunc(logsPath, LogDataHandler())

	//log.Infof("http control server listening on %s", controlAddr)
	//if err := http.ListenAndServe(controlAddr, nil); err != nil {
	//	log.Error(err.Error())
	//	os.Exit(1)
	//}

	if err := http.ListenAndServe(logsPort, nil); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func ServeProxy(store *MetricStore) {
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
}

func HandleIncomingMetrics(store *MetricStore, conn net.Conn) {
	defer conn.Close()
	lines := bufio.NewScanner(conn)
	for lines.Scan() {
		if len(lines.Text()) == 0 {
			continue
		}
		str := lines.Text()
		log.Infof(fmt.Sprintf("Print from tcp listener -- %s", str))

		metric, err := ParseMetric(str)
		if err != nil {
			log.Error(err.Error())
			log.Error(lines.Text())
			store.LogBadMetric(lines.Text())
			continue
		}
		if metric == nil { // we got a histogram
			continue
		}
		if len(metric.Tags) > 20 {
			log.Error(fmt.Sprintf("[WF-410: Too many point tags (%d, max 20):", len(metric.Tags)))
			continue
		}
		log.Debugf("%#v", metric)
		store.LogMetric(metric)
	}
	if err := lines.Err(); err != nil {
		log.Error(err.Error())
	}
	return
}

func LogDataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != logsPath {
			http.NotFound(w, req)
			return
		}

		//lines := bufio.NewScanner(req.Body)
		//defer req.Body.Close()
		//for lines.Scan() {
		//	if len(lines.Bytes()) == 0 {
		//		continue
		//	}
		//	log.Info(lines.Text())
		//
		//}

		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		log.Info(string(b))
		// Verify that the input is in JSON array format
		if VerifyJsonArray(string(b)) {
			log.Info("logs are in json_array format")
		} else {
			log.Info("logs are not in json_array format")
			return
		}

		//linesErr := lines.Err()
		//if linesErr != nil {
		//	w.WriteHeader(http.StatusBadRequest)
		//	ioErr := json.NewEncoder(w).Encode(linesErr.Error())
		//	if ioErr != nil {
		//		log.Error(ioErr.Error())
		//	}
		//	return
		//}
		w.WriteHeader(http.StatusOK)
		// Read the incoming bytes
		//b := make([]byte, req.ContentLength)
		//_, err := r.Body.Read(b)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}

		// Print the input to the console
		//fmt.Println("Print from HTTP listener --" + string(b))

		//if r.Method != http.MethodPost {
		//	w.WriteHeader(http.StatusMethodNotAllowed)
		//	log.Errorf("expected method %s but got %s", http.MethodPost, r.Method)
		//	return
		//}
	}
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
			log.Errorf("expected method %s but got %s", http.MethodPost, req.Method)
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
		var excludedMetrics []*Metric
		lines := bufio.NewScanner(req.Body)
		defer req.Body.Close()
		for lines.Scan() {
			if len(lines.Bytes()) == 0 {
				continue
			}
			var err error
			if lines.Bytes()[0] == '~' {
				var excludedMetric *Metric
				excludedMetric, err = decodeMetric(lines.Bytes()[1:])
				excludedMetrics = append(excludedMetrics, excludedMetric)
			} else {
				var expectedMetric *Metric
				expectedMetric, err = decodeMetric(lines.Bytes())
				expectedMetrics = append(expectedMetrics, expectedMetric)
			}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(err.Error())
				if err != nil {
					log.Error(err.Error())
				}
				return
			}
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
		linesErr = json.NewEncoder(w).Encode(DiffMetrics(expectedMetrics, excludedMetrics, store.Metrics()))
		if linesErr != nil {
			log.Error(linesErr.Error())
		}
	}
}

func decodeMetric(bytes []byte) (*Metric, error) {
	var metric *Metric
	err := json.Unmarshal(bytes, &metric)
	return metric, err
}
