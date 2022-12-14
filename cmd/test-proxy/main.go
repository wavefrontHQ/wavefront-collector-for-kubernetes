package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
)

var proxyAddr = ":7777"
var controlAddr = ":8888"
var logLevel = log.InfoLevel.String()

func init() {
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, "host and port for the test \"wavefront proxy\" to listen on")
	flag.StringVar(&controlAddr, "control", controlAddr, "host and port for the http control server to listen on")
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

	store := NewMetricStore()

	go ServeProxy(store)

	http.HandleFunc("/metrics", DumpMetricsHandler(store))
	http.HandleFunc("/metrics/diff", DiffMetricsHandler(store))
	// Based on logs already sent, perform checks on store logs
	// Start by supporting POST parameter expected_log_format
	http.HandleFunc("/logs/assert", LogAssertionHandler())
	// NOTE: these handler functions attach to the control HTTP server, NOT the TCP server that actually receives data

	log.Infof("http control server listening on %s", controlAddr)
	if err := http.ListenAndServe(controlAddr, nil); err != nil {
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

func getLineType(string2 string) {

}

func HandleIncomingMetrics(store *MetricStore, conn net.Conn) {
	defer conn.Close()
	lines := bufio.NewScanner(conn)

	// TODO: Read entire http/tcp request
	// TODO: Parse the body between metric and log, using POST prefix: POST /logs/json_array?f=logs_json_arr or POST /logs/json_lines?f=logs_json_lines <- (maybe?)
	// NOTE start with POST /logs/json_array?f=logs_json_arr just to get it working
	var b = make([]byte, 1024)
	_, err := conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	// NOTE  store logs request body in order of how we receive them, then later we can perform verification on the rest of them
	log.Infof("******* string(b) in HandleIncomingMetrics: %s", string(b))

	// METRICS:
	// \"~wavefront.kubernetes.collector.version\" 1.13 source=\"kind-control-plane\" \"cluste │
	//│ r\"=\"test-if-can-have-no-proxy\" \"stats_prefix\"=\"kubernetes.\" \"installation_method\"=\"operator\"\n\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x │
	//│ 00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\ │
	//│ x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 │
	//│ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0 │
	//│ 0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x │
	//│ 00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\ │
	//│ x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 │
	//│ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0 │
	//│ 0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x │
	//│ 00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\ │
	//│ x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 │
	//│ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0 │
	//│ 0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x │
	//│ 00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\ │
	//│ x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 │
	//│ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0 │
	//│ 0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x │
	//│ 00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\ │
	//│ x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 │
	//│ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0 │
	//│ 0"

	// LOGS:
	// ******* string(b) in HandleIncomingMetrics: POST /logs/json_array?f=logs_json_arr HTTP/1.1\r\nAccept-Encoding: gzip;q=1.0,deflate;q │
	// │ =0.6,identity;q=0.3\r\nAccept: */*\r\nUser-Agent: Ruby\r\nContent-Type: application/json\r\nHost: test-proxy.observability-fake-proxy.svc.cluster.local:2878\r\nContent-Length: │
	// │  33804\r\n\r\n[{\"stream\":\"stderr\",\"character\":\"F\",\"message\":\"I1214 22:05:17.574634       1 main.go:223] Handling node with IPs: map[172.18.0.2:{}]\",\"service\":\"n │
	// │ one\",\"application\":\"none\",\"source\":\"kind-control-plane\",\"cluster\":\"test-if-can-have-no-proxy\",\"timestamp\":1671055517575,\"pod_name\":\"kindnet-gzsvn\",\"contain │
	// │ er_name\":\"kindnet-cni\",\"namespace_name\":\"kube-system\",\"pod_id\":\"2de6a61a-6eab-46f1-ba31-47d5fd835fe5\",\"container_id\":\"e36566a15fe7a20cbe5e96686d18961b35fd5b4259a │
	// │ 6e2bf46a2ba5ff463c676\"}\n,{\"stream\":\"stderr\",\"character\":\"F\",\"message\":\"I1214 22:05:17.574727       1 main.go:227] handling current node\",\"service\":\"none\",\"a │
	// │ pplication\":\"none\",\"source\":\"kind-control-plane\",\"cluster\":\"test-if-can-have-no-proxy\",\"timestamp\":1671055517575,\"pod_name\":\"kindnet-gzsvn\",\"conta"

	// We might filter logs first with "POST /logs/json_array?f=logs_json_arr" at first and parse the rest incoming traffic with lines.Scan as Metric
	for lines.Scan() {
		if len(lines.Text()) == 0 {
			continue
		}
		str := lines.Text()
		//log.Infof("*********** str = %s", str)
		//if strings.Contains(str, "[") {
		//	log.Infof("**************** supposedly a JSON array: %s", str)
		//	break
		//}

		metric, err := ParseMetric(str)
		if err != nil {
			//log.Error(err.Error())
			//log.Error(lines.Text())
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
		//log.Debugf("%#v", metric)
		store.LogMetric(metric)
	}
	if err := lines.Err(); err != nil {
		//log.Error(err.Error())
	}
	return
}

func LogJsonArrayHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Infof("******* req received in LogJsonArrayHandler: '%+v'", req)

		//if req.URL.Path != logsPath {
		//	http.NotFound(w, req)
		//	return
		//}

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

		log.Infof("******* string(b): %s", string(b))
		// Verify that the input is in JSON array format
		// TODO: use VerifyJsonArray on the body of the tcp request and store

		if VerifyJsonArray(string(b)) {
			log.Info("******* logs are in json_array format")
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Info("logs are not in json_array format")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func LogAssertionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Infof("******* req received in LogAssertionHandler: '%+v'", req)

		//if req.URL.Path != logsPath {
		//	http.NotFound(w, req)
		//	return
		//}

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

		log.Infof("******* string(b): %s", string(b))
		// Verify that the input is in JSON array format

		if VerifyJsonArray(string(b)) {
			log.Info("******* logs are in json_array format")
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Info("logs are not in json_array format")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func LogJsonLinesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//if req.URL.Path != logsPath {
		//	http.NotFound(w, req)
		//	return
		//}

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

		if VerifyJsonLines(string(b)) {
			log.Info("logs are in json_lines format")
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Info("logs are not in json_lines format")
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}

func DumpMetricsHandler(store *MetricStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Infof("******* req received in DumpMetricsHandler: '%+v'", req)
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
