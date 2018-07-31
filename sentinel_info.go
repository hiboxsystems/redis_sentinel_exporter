package main

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

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

func (m *Master) StatusAsFloat64() float64 {
	if m.Status == "ok" {
		return float64(1)
	}
	return float64(0)
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
