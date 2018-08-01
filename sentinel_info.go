package main

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	masterMetricMap = map[string]string{
		"name":      "name",
		"status":    "status",
		"address":   "address",
		"slaves":    "slaves",
		"sentinels": "sentinels",
	}
)

type SentinelInfo struct {
	Metrics map[string]interface{}
	Masters []*Master
}

type Master struct {
	Metrics map[string]interface{}
}

// Format:
// name=mymaster,status=ok,address=172.17.8.101:6379,slaves=2,sentinels=3
func pasreMasterInfo(info string) *Master {
	split := strings.Split(info, ",")
	m := &Master{
		Metrics: make(map[string]interface{}),
	}
	for _, keyPair := range split {
		s := strings.Split(keyPair, "=")
		if len(s) != 2 {
			continue
		}
		fieldKey := s[0]
		fieldValue := s[1]
		for metricOriginalName, metricName := range masterMetricMap {
			if metricOriginalName != fieldKey {
				continue
			}
			m.Metrics[metricName] = parseValue(fieldValue)
		}
	}
	return m
}

func parseValue(value string) interface{} {
	if value == "ok" {
		return float64(1)
	} else if value == "fail" || value == "down" {
		return float64(0)
	} else if val, err := strconv.ParseFloat(value, 64); err == nil {
		return val
	}
	return value
}

func parseInfo(info string, keys []string, includeMasters bool) *SentinelInfo {
	lines := strings.Split(info, "\r\n")
	i := &SentinelInfo{
		Metrics: make(map[string]interface{}),
	}
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
		if strings.HasPrefix(fieldKey, "master") && includeMasters {
			master := pasreMasterInfo(fieldValue)
			i.Masters = append(i.Masters, master)
			continue
		}
		for _, key := range keys {
			if key != fieldKey {
				continue
			}
			i.Metrics[key] = parseValue(fieldValue)
		}
	}
	return i
}
