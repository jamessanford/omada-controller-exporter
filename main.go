package main

// Export wireless station metrics from an Omada WiFi Controller

import (
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/jamessanford/omada-controller-exporter/collector"
	"github.com/jamessanford/omada-controller-exporter/omada"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "", "path to YAML config")
	httpAddr   = flag.String("http", ":6779", "listen on this address")
)

const usageMessage = `Connect to a TP-Link Omada Controller and expose metrics.

Usage:
    omada-controller-exporter [-config <file>] [-http <listen address>]

Configuration:
    Use either -config <file> against a YAML file like this:

path: https://192.168.255.123:8043/
user: admin
pass: foo
secure: false

    Or configure via environment variables:

OMADA_PATH=https://192.168.255.123:8043/
OMADA_USER=admin
OMADA_PASS=foo
OMADA_SECURE=false

Options:
`

func usage() {
	os.Stderr.Write([]byte(usageMessage))
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	config, err := omada.ParseConfig(*configFile)
	if err != nil {
		panic(err)
	}

	controller, err := omada.NewClient(logger, config)
	if err != nil {
		panic(err)
	}

	prometheus.MustRegister(collector.NewOmadaCollector(logger, controller))

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "omada-controller-exporter")
	})
	http.Handle("/metrics", promhttp.Handler())

	logger.Info(
		"listening",
		zap.String("address", *httpAddr),
	)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		logger.Fatal("ListenAndServe", zap.Error(err))
	}
}
