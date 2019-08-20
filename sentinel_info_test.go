package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestParseValue(t *testing.T) {
	cases := map[string]interface{}{
		"ok":     1.0,
		"down":   0.0,
		"fail":   0.0,
		"13.0":   13.0,
		"foobar": "foobar",
	}
	for in, out := range cases {
		if val := ParseValue(in); val != out {
			t.Errorf("Must be %v, but got %v", out, val)
		}
	}
}

func TestPasreMasterInfo(t *testing.T) {
	masterA := PasreMasterInfo("foobar,foo=bar,name=mymaster,status=ok,address=127.0.0.1:6379,slaves=2,sentinels=3")
	if masterA == nil {
		t.Error("Must be Master, but got nil")
	}
	if masterA.Metrics["name"].(string) != "mymaster" {
		t.Errorf("Must be mymaster, but got %v", masterA.Metrics["name"])
	}
	if _, ok := masterA.Metrics["foobar"]; ok {
		t.Error("Must be false, but got true")
	}
	if _, ok := masterA.Metrics["foo"]; ok {
		t.Error("Must be false, but got true")
	}
	if masterA.Metrics["status"].(float64) != 1.0 {
		t.Errorf("Must be 1.0, but got %v", masterA.Metrics["status"])
	}
}

func TestParseInfo(t *testing.T) {
	tests := []struct {
		filename string
		master   bool
	}{
		{
			filename: "test_data/case-1",
			master:   true,
		},
		{
			filename: "test_data/case-2",
		},
	}
	for _, test := range tests {
		b, err := ioutil.ReadFile(test.filename)
		if err != nil {
			t.Fatal(err)
		}
		b = bytes.Replace(b, []byte("\n"), []byte("\r\n"), -1)
		metricRequiredKeys := metricBuildInfo
		for metricName := range metricMap {
			metricRequiredKeys = append(metricRequiredKeys, metricName)
		}
		si := ParseInfo(string(b), metricRequiredKeys, test.master)
		if si == nil {
			t.Error("Must be an object, but got nil")
		}
	}
}
