package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, out, ParseValue(in))
	}
}

func TestParseMasterInfo(t *testing.T) {
	masterA := ParseMasterInfo("foobar,foo=bar,name=mymaster,status=ok,address=127.0.0.1:6379,slaves=2,sentinels=3")
	assert.NotNil(t, masterA)
	assert.Equal(t, "mymaster", masterA.Metrics["name"].(string))

	_, ok := masterA.Metrics["foobar"]
	assert.False(t, ok)

	_, ok = masterA.Metrics["foo"]
	assert.False(t, ok)
	assert.Equal(t, 1.0, masterA.Metrics["status"].(float64))
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
		assert.NotNil(t, si)
	}
}
