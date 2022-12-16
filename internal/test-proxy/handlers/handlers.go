package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/test-proxy/logs"
	metrics2 "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/test-proxy/metrics"
	"io"
	"net"
	"net/http"
)

func LogJsonArrayHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//log.Infof("******* req received in LogJsonArrayHandler: '%+v'", req)

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

		if logs.VerifyJsonArray(string(b)) {
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

		if logs.VerifyJsonArray(string(b)) {
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

		if logs.VerifyJsonLines(string(b)) {
			log.Info("logs are in json_lines format")
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Info("logs are not in json_lines format")
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}

func HandleIncomingMetrics(store *metrics2.MetricStore, conn net.Conn) {
	defer conn.Close()
	lines := bufio.NewScanner(conn)

	for lines.Scan() {
		if len(lines.Text()) == 0 {
			continue
		}
		str := lines.Text()

		metric, err := metrics2.ParseMetric(str)
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

func DumpMetricsHandler(store *metrics2.MetricStore) http.HandlerFunc {
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

func DiffMetricsHandler(store *metrics2.MetricStore) http.HandlerFunc {
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

		var expectedMetrics []*metrics2.Metric
		var excludedMetrics []*metrics2.Metric
		lines := bufio.NewScanner(req.Body)
		defer req.Body.Close()

		for lines.Scan() {
			if len(lines.Bytes()) == 0 {
				continue
			}
			var err error
			if lines.Bytes()[0] == '~' {
				var excludedMetric *metrics2.Metric
				excludedMetric, err = decodeMetric(lines.Bytes()[1:])
				excludedMetrics = append(excludedMetrics, excludedMetric)
			} else {
				var expectedMetric *metrics2.Metric
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

		linesErr = json.NewEncoder(w).Encode(metrics2.DiffMetrics(expectedMetrics, excludedMetrics, store.Metrics()))
		if linesErr != nil {
			log.Error(linesErr.Error())
		}
	}
}

func decodeMetric(bytes []byte) (*metrics2.Metric, error) {
	var metric *metrics2.Metric
	err := json.Unmarshal(bytes, &metric)
	return metric, err
}
