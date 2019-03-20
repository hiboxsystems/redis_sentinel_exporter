package main

import (
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	metricBuildInfo = []string{
		"redis_version",
		"redis_build_id",
		"redis_mode",
	}
	metricMap = map[string]string{
		// Server
		"uptime_in_seconds": "uptime_in_seconds",
		"process_id":        "process_id",
		// Clients
		"connected_clients":          "connected_clients",
		"client_longest_output_list": "client_longest_output_list",
		"client_biggest_input_buf":   "client_biggest_input_buf",
		"blocked_clients":            "blocked_clients",
		// Stats
		"total_connections_received": "connections_received_total",
		"total_commands_processed":   "commands_processed_total",
		"instantaneous_ops_per_sec":  "instantaneous_ops_per_sec",
		"total_net_input_bytes":      "net_input_bytes_total",
		"total_net_output_bytes":     "net_output_bytes_total",
		"instantaneous_input_kbps":   "instantaneous_input_kbps",
		"instantaneous_output_kbps":  "instantaneous_output_kbps",
		"rejected_connections":       "rejected_connections_total",
		"expired_keys":               "expired_keys_total",
		"evicted_keys":               "evicted_keys_total",
		"keyspace_hits":              "keyspace_hits_total",
		"keyspace_misses":            "keyspace_misses_total",
		"pubsub_channels":            "pubsub_channels",
		"pubsub_patterns":            "pubsub_patterns",
		"latest_fork_usec":           "latest_fork_usec",
		// CPU
		"used_cpu_sys":           "used_cpu_sys",
		"used_cpu_user":          "used_cpu_user",
		"used_cpu_sys_children":  "used_cpu_sys_children",
		"used_cpu_user_children": "used_cpu_user_children",
		// Sentinel
		"sentinel_masters":                "masters",
		"sentinel_tilt":                   "tilt",
		"sentinel_running_scripts":        "running_scripts",
		"sentinel_scripts_queue_length":   "scripts_queue_length",
		"sentinel_simulate_failure_flags": "simulate_failure_flags",
	}
)

type Exporter struct {
	addr         string
	namespace    string
	metrics      map[string]*prometheus.GaugeVec
	duration     prometheus.Gauge
	scrapeErrors prometheus.Gauge
	totalScrapes prometheus.Counter
}

func NewRedisSentinelExporter(addr, namespace string) *Exporter {
	e := &Exporter{
		addr:      addr,
		namespace: namespace,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exporter_last_scrape_duration_seconds",
			Help:      "The last scrape duration.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total redis scrapes.",
		}),
		scrapeErrors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exporter_last_scrape_error",
			Help:      "The last scrape error status.",
		}),
	}
	e.initGauges()
	return e
}

func (e *Exporter) initGauges() {
	e.metrics = map[string]*prometheus.GaugeVec{}

	e.metrics["info"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "info",
		Help:      "Information about Sentinel",
	}, []string{"version", "build_id", "mode"})

	// Masters
	e.metrics["master_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_status",
		Help:      "Status of master",
	}, []string{"name", "address"})
	e.metrics["master_slaves"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_slaves",
		Help:      "Slaves of master",
	}, []string{"name", "address"})
	e.metrics["master_sentinels"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_sentinels",
		Help:      "Sentinels of master",
	}, []string{"name", "address"})

	// All other metrics
	for metricOrigName, metricOutName := range metricMap {
		e.metrics[metricOrigName] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: e.namespace,
			Name:      metricOutName,
			Help:      metricOutName,
		}, []string{})
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.duration.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.scrapeErrors.Desc()

	for _, m := range e.metrics {
		m.Describe(ch)
	}
}

func (e *Exporter) scrapeInfo() (string, error) {
	options := []redis.DialOption{
		redis.DialConnectTimeout(5 * time.Second),
		redis.DialReadTimeout(5 * time.Second),
		redis.DialWriteTimeout(5 * time.Second),
	}

	logrus.Debugf("Trying DialURL(): %s", e.addr)
	c, err := redis.DialURL(e.addr, options...)

	if err != nil {
		logrus.Debugf("DialURL() failed, err: %s", err)
		if frags := strings.Split(e.addr, "://"); len(frags) == 2 {
			logrus.Debugf("Trying: Dial(): %s %s", frags[0], frags[1])
			c, err = redis.Dial(frags[0], frags[1], options...)
		} else {
			logrus.Debugf("Trying: Dial(): tcp %s", e.addr)
			c, err = redis.Dial("tcp", e.addr, options...)
		}
	}

	if err != nil {
		logrus.Debugf("aborting for addr: %s - redis sentinel err: %s", e.addr, err)
		return "", err
	}

	defer c.Close()
	logrus.Debugf("connected to: %s", e.addr)

	body, err := redis.String(c.Do("info"))
	if err != nil {
		logrus.Debugf("cannot execute command info: %v", err)
		return "", err
	}

	return body, nil
}

func (e *Exporter) setMetrics(i *SentinelInfo) {
	for metricName, gauge := range e.metrics {
		switch metricName {
		case "info":
			gauge.WithLabelValues(
				i.Metrics["redis_version"].(string),
				i.Metrics["redis_build_id"].(string),
				i.Metrics["redis_mode"].(string),
			).Set(float64(1))
		case "master_status", "master_slaves", "master_sentinels":
			gauge.Reset()
			metricType := strings.TrimPrefix(metricName, "master_")
			for _, m := range i.Masters {
				gauge.WithLabelValues(
					m.Metrics["name"].(string),
					m.Metrics["address"].(string),
				).Set(m.Metrics[metricType].(float64))
			}
		default:
			if _, ok := i.Metrics[metricName]; ok {
				gauge.WithLabelValues().Set(i.Metrics[metricName].(float64))
			}
		}
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var errorCount int
	now := time.Now().UnixNano()
	e.totalScrapes.Inc()

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.scrapeErrors

	infoRaw, err := e.scrapeInfo()
	metricRequiredKeys := metricBuildInfo
	for metricName := range metricMap {
		metricRequiredKeys = append(metricRequiredKeys, metricName)
	}
	if err != nil {
		errorCount++
	} else {
		sentinelInfo := ParseInfo(infoRaw, metricRequiredKeys, true)
		e.setMetrics(sentinelInfo)
	}

	e.scrapeErrors.Set(float64(errorCount))
	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)

	for _, m := range e.metrics {
		m.Collect(ch)
	}
}
