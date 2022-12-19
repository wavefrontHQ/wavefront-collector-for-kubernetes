package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/logs"
	metrics2 "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/testproxy/metrics"
)

func LogJsonArrayHandler(store *logs.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		store.SetReceivedWithValidFormat(logs.VerifyJsonArray(string(b)))

		w.WriteHeader(http.StatusOK)
	}
}

func LogJsonLinesHandler(store *logs.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		store.SetReceivedWithValidFormat(logs.VerifyJsonLines(string(b)))

		w.WriteHeader(http.StatusOK)
	}
}

func LogAssertionHandler(store *logs.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		output, err := store.ToJSON()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Unable to marshal log test store object: %s", err.Error())))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(output)
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
