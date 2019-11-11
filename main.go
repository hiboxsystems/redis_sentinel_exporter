package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

var (
	listenAddress    = flag.String("web.listen-address", ":9355", "Address to listen on for web interface and telemetry.")
	metricPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	sentinelAddr     = flag.String("sentinel.addr", "redis://127.0.0.1:26379", "Redis Sentinel host:port")
	sentinelPassword = flag.String("sentinel.password", "", "Redis Sentinel password (optional)")
	isDebug          = flag.Bool("debug", false, "Output verbose debug information")
	logFormat        = flag.String("log-format", "txt", "Log format, valid options are txt and json")
	namespace        = flag.String("namespace", "redis_sentinel", "Namespace for metrics")
	versionPrint     = flag.Bool("version", false, "Prints version and exit")
)

func main() {
	flag.Parse()

	if *versionPrint {
		fmt.Println(version.Print("redis sentinel exporter"))
		os.Exit(0)
	}

	switch *logFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.Infof("Starting Redis Sentinel Exporter %s...", version.Version)

	if *isDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Enabling debug output")
	}

	if len(*sentinelAddr) == 0 {
		logrus.Fatal("Must specify a non-empty sentinel.addr")
	}

	exp := NewRedisSentinelExporter(*sentinelAddr, *namespace, *sentinelPassword)

	prometheus.MustRegister(exp)
	http.Handle(*metricPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Redis Sentinel Exporter ` + version.Version + `</title></head>
			<body>
			<h1>Redis Sentinel Exporter ` + version.Version + `</h1>
			<p><a href='` + *metricPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	logrus.Printf("Providing metrics at %s%s", *listenAddress, *metricPath)
	logrus.Fatal(http.ListenAndServe(*listenAddress, nil))
}
