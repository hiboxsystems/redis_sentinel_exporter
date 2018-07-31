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

type SentinelInfo struct {
	// Server
	Version         string
	BuildID         string
	Mode            string
	UptimeInSeconds float64
	// Sentinel
	Masters            float64
	Tilt               float64
	RunningScripts     float64
	ScriptsQueueLength float64
	MastersList        []*Master
}

type Master struct {
	ID        int
	Name      string
	Status    string
	Address   string
	Slaves    float64
	Sentinels float64
}

func pasreMasterInfo(info string) *Master {
	split := strings.Split(info, ",")
	m := &Master{}
	for _, keyPair := range split {
		s := strings.Split(keyPair, "=")
		if len(s) != 2 {
			continue
		}
		fieldKey := s[0]
		fieldValue := s[1]
		switch fieldKey {
		case "name":
			m.Name = fieldValue
		case "status":
			m.Status = fieldValue
		case "address":
			m.Address = fieldValue
		case "slaves":
			slaves, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				m.Slaves = slaves
			}
		case "sentinels":
			sentinels, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				m.Sentinels = sentinels
			}
		}
	}
	return m
}

func parseInfo(info string) *SentinelInfo {
	lines := strings.Split(info, "\r\n")
	i := &SentinelInfo{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		logrus.Debugf("info: %s", line)
		split := strings.SplitN(line, ":", 2)
		if len(split) < 2 {
			continue
		}
		fieldKey := split[0]
		fieldValue := split[1]
		if strings.HasPrefix(fieldKey, "master") {
			// name=mymaster,status=ok,address=172.17.8.101:6379,slaves=2,sentinels=3
			master := pasreMasterInfo(fieldValue)
			i.MastersList = append(i.MastersList, master)
			continue
		}
		switch fieldKey {
		case "redis_version":
			i.Version = fieldValue
		case "redis_build_id":
			i.BuildID = fieldValue
		case "redis_mode":
			i.Mode = fieldValue
		case "uptime_in_seconds":
			uptime, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				i.UptimeInSeconds = uptime
			}
		case "sentinel_masters":
			masters, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				i.Masters = masters
			}
		case "sentinel_tilt":
			tilt, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				i.Tilt = tilt
			}
		case "sentinel_running_scripts":
			scripts, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				i.RunningScripts = scripts
			}
		case "sentinel_scripts_queue_length":
			scripts, err := strconv.ParseFloat(fieldValue, 64)
			if err == nil {
				i.ScriptsQueueLength = scripts
			}
		}
	}
	return i
}

func (m *Master) StatusAsFloat64() float64 {
	if m.Status == "ok" {
		return float64(1)
	}
	return float64(0)
}

func (e *Exporter) setMetrics(info *SentinelInfo) {
	e.metrics["info"].WithLabelValues(info.Version, info.BuildID, info.Mode).Set(float64(1))
	e.metrics["masters"].WithLabelValues().Set(info.Masters)
	e.metrics["tilt"].WithLabelValues().Set(info.Tilt)
	e.metrics["running_scripts"].WithLabelValues().Set(info.RunningScripts)
	e.metrics["scripts_queue_length"].WithLabelValues().Set(info.ScriptsQueueLength)
	e.metrics["uptime_seconds"].WithLabelValues().Set(info.UptimeInSeconds)
	for _, master := range info.MastersList {
		e.metrics["master_status"].WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).Set(master.StatusAsFloat64())
		e.metrics["master_slaves"].WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).Set(master.Slaves)
		e.metrics["master_sentinels"].WithLabelValues(strconv.Itoa(master.ID), master.Name, master.Address).Set(master.Sentinels)
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
