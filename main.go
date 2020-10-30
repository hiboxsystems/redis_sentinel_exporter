package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

var (
	isDebug              = flag.Bool("debug", GetBoolEnv("DEBUG", false), "Output verbose debug information.")
	listenAddress        = flag.String("web.listen-address", GetStringEnv("LISTEN_ADDRESS", ":9355"), "Address to listen on for web interface and telemetry.")
	logFormat            = flag.String("log-format", GetStringEnv("LOG_FORMAT", "txt"), "Log format, valid options are txt and json.")
	metricPath           = flag.String("web.telemetry-path", GetStringEnv("TELEMETRY_PATH", "/metrics"), "Path under which to expose metrics.")
	namespace            = flag.String("namespace", GetStringEnv("NAMESPACE", "redis_sentinel"), "Namespace for metrics.")
	sentinelAddr         = flag.String("sentinel.addr", GetStringEnv("SENTINEL_ADDR", "redis://127.0.0.1:26379"), "Redis Sentinel host:port.")
	sentinelPassword     = flag.String("sentinel.password", GetStringEnv("SENTINEL_PASSWORD", ""), "Redis Sentinel password (optional).")
	sentinelPasswordFile = flag.String("sentinel.password-file", GetStringEnv("SENTINEL_PASSWORD_FILE", ""), "Path to Redis Sentinel password file (optional).")
	versionPrint         = flag.Bool("version", false, "Prints version and exit.")
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

	password := *sentinelPassword
	if len(*sentinelPasswordFile) > 0 {
		body, err := ioutil.ReadFile(*sentinelPasswordFile)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load Redis Sentinel password file")
		}
		password = strings.TrimSpace(string(body))
	}

	exporter := NewRedisSentinelExporter(*sentinelAddr, *namespace, password)

	prometheus.MustRegister(exporter)

	// Deprecated
	prometheus.MustRegister(version.NewCollector(*namespace))
	versionNamespace := *namespace + "_exporter"
	prometheus.MustRegister(version.NewCollector(versionNamespace))

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
