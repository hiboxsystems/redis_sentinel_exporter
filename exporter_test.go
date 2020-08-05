package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisSentinelExporter(t *testing.T) {
	e := NewRedisSentinelExporter("127.0.0.1:8080", "ns", "")
	assert.NotNil(t, e)
}
