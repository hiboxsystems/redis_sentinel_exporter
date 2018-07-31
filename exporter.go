package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
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
	e.metrics["masters"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "masters",
		Help:      "Total masters",
	}, []string{})
	e.metrics["tilt"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "tilt",
		Help:      "Tilt value",
	}, []string{})
	e.metrics["running_scripts"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "running_scripts",
		Help:      "Number of running scripts",
	}, []string{})
	e.metrics["scripts_queue_length"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "scripts_queue_length",
		Help:      "Length of scripts queue",
	}, []string{})
	e.metrics["master_status"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_status",
		Help:      "Status of master",
	}, []string{"id", "name", "address"})
	e.metrics["master_slaves"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_slaves",
		Help:      "Slaves of master",
	}, []string{"id", "name", "address"})
	e.metrics["master_sentinels"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "master_sentinels",
		Help:      "Sentinels of master",
	}, []string{"id", "name", "address"})
	e.metrics["info"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "info",
		Help:      "Information about Sentinel",
	}, []string{"version", "build_id", "mode"})
	e.metrics["uptime_seconds"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: e.namespace,
		Name:      "uptime_seconds",
		Help:      "Sentinel uptime in seconds",
	}, []string{})
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

func (e *Exporter) setMetrics(info *SentinelInfo) {
	e.metrics["info"].WithLabelValues(info.Version, info.BuildID, info.Mode).Set(float64(1))
	e.metrics["masters"].WithLabelValues().Set(info.Masters)
	e.metrics["tilt"].WithLabelValues().Set(info.Tilt)
	e.metrics["running_scripts"].WithLabelValues().Set(info.RunningScripts)
	e.metrics["scripts_queue_length"].WithLabelValues().Set(info.ScriptsQueueLength)
	e.metrics["uptime_seconds"].WithLabelValues().Set(info.UptimeInSeconds)

	for _, master := range info.MastersList {
		e.metrics["master_status"].
			WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).
			Set(master.StatusAsFloat64())
		e.metrics["master_slaves"].
			WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).
			Set(master.Slaves)
		e.metrics["master_sentinels"].
			WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).
			Set(master.Sentinels)
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
	if err != nil {
		errorCount++
	} else {
		sentinelInfo := parseInfo(infoRaw)
		e.setMetrics(sentinelInfo)
	}

	e.scrapeErrors.Set(float64(errorCount))
	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)

	for _, m := range e.metrics {
		m.Collect(ch)
	}
}
