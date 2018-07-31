package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	listenAddress = flag.String("web.listen-address", ":9355", "Address to listen on for web interface and telemetry.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sentinelAddr  = flag.String("sentinel.addr", "redis://127.0.0.1:26379", "Redis Sentinel host:port")
	isDebug       = flag.Bool("debug", false, "Output verbose debug information")
	logFormat     = flag.String("log-format", "txt", "Log format, valid options are txt and json")
	namespace     = flag.String("namespace", "redis_sentinel", "Namespace for metrics")
)

func main() {
	flag.Parse()

	switch *logFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if *isDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Enabling debug output")
	}

	if len(*sentinelAddr) == 0 {
		logrus.Fatal("Must specify a non-empty sentinel.addr")
	}

	exp := NewRedisSentinelExporter(*sentinelAddr, *namespace)

	prometheus.MustRegister(exp)
	http.Handle(*metricPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Redis Sentine Exporter</title></head>
			<body>
			<h1>Redis Sentinel Exporter</h1>
			<p><a href='` + *metricPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	logrus.Printf("Providing metrics at %s%s", *listenAddress, *metricPath)
	logrus.Fatal(http.ListenAndServe(*listenAddress, nil))
}
